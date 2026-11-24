package main // this is an executable program

import ( // standard library packages used by this program
	"crypto/tls"         // describing TLS connection details in the trace logs
	"fmt"                // building formatted error messages and urls
	"html"               // decoding HTML entities such as &amp; inside links
	"io"                 // reading response bodies and copying them to files
	"log"                // fatal errors (its output is routed through slog after slog.SetDefault)
	"log/slog"           // leveled structured logging, used here at DEBUG level
	"net/http"           // making HTTP requests
	"net/http/httptrace" // tracing DNS, connection, TLS and timing events of each request
	"net/url"            // parsing and resolving urls
	"os"                 // creating folders and files, checking if files exist
	"path"               // getting the last element of a url path
	"path/filepath"      // building local file paths for the current operating system
	"regexp"             // finding PDF links inside the page HTML
	"runtime"            // reporting Go version, operating system and architecture
	"strings"            // trimming and replacing text
	"time"               // timeouts, delays and measuring durations
) // end of imports

const ( // fixed settings for the program
	outputDirectoryName  = "PDFs"                                                                                                        // folder where the PDFs are saved
	firstPageNumber      = 1                                                                                                             // first download page to visit
	lastPageNumber       = 10                                                                                                            // last download page to visit
	downloadPageBaseUrl  = "https://myklaticrete.com/downloads/"                                                                         // url of page 1
	requestTimeout       = 60 * time.Second                                                                                              // maximum time for one HTTP request
	delayBetweenRequests = 500 * time.Millisecond                                                                                        // pause between downloads to be polite to the server
	maximumRedirectCount = 10                                                                                                            // stop following redirects after this many hops
	browserUserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36" // pretend to be a normal browser
) // end of constants

var portableDocumentFileLinkPattern = regexp.MustCompile(`(?i)(?:data-pdf-url|href)\s*=\s*["']([^"']+?\.pdf(?:\?[^"']*)?)["']`) // matches data-pdf-url="...pdf" (and href="...pdf") attributes, ignoring letter case

func main() { // program starts here
	programStartTime := time.Now() // remember when the program started so total time can be reported

	debugLogHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ // create a text handler that prints to the terminal
		Level:     slog.LevelDebug, // show every message, including DEBUG
		AddSource: true,            // include the source file and line number in every message
	}) // end of handler options
	slog.SetDefault(slog.New(debugLogHandler)) // make this the default logger (the log package now routes through it too)

	slog.Debug("program started", "goVersion", runtime.Version(), "operatingSystem", runtime.GOOS, "architecture", runtime.GOARCH, "processIdentifier", os.Getpid()) // show the runtime environment

	workingDirectory, getWorkingDirectoryError := os.Getwd() // find the folder the program is running from
	if getWorkingDirectoryError != nil {                     // check whether that failed
		slog.Warn("could not determine working directory", "error", getWorkingDirectoryError) // warn but keep going
	} else { // the working directory is known
		slog.Debug("working directory", "path", workingDirectory) // show where files will be saved
	} // end of working directory check

	slog.Debug("configuration", "outputDirectoryName", outputDirectoryName, "firstPageNumber", firstPageNumber, "lastPageNumber", lastPageNumber, "downloadPageBaseUrl", downloadPageBaseUrl, "requestTimeout", requestTimeout, "delayBetweenRequests", delayBetweenRequests, "maximumRedirectCount", maximumRedirectCount) // show every setting

	slog.Debug("creating output directory if missing", "directory", outputDirectoryName) // announce the folder step
	makeDirectoryError := os.MkdirAll(outputDirectoryName, 0o755)                        // create the PDFs folder if it does not exist
	if makeDirectoryError != nil {                                                       // check whether the folder could not be created
		log.Fatalf("could not create output directory %q: %v", outputDirectoryName, makeDirectoryError) // print the error and stop
	} // end of folder error check

	existingDirectoryEntryList, readDirectoryError := os.ReadDir(outputDirectoryName) // list what is already in the folder
	if readDirectoryError != nil {                                                    // check whether listing failed
		slog.Warn("could not list output directory", "error", readDirectoryError) // warn but keep going
	} else { // listing worked
		slog.Debug("output directory ready", "directory", outputDirectoryName, "existingEntryCount", len(existingDirectoryEntryList)) // show how many entries already exist
	} // end of directory listing check

	httpClient := &http.Client{ // one reusable HTTP client
		Timeout: requestTimeout, // maximum time for one request
		CheckRedirect: func(nextRequest *http.Request, previousRequestList []*http.Request) error { // called before following every redirect
			slog.Debug("following redirect", "hopNumber", len(previousRequestList), "from", previousRequestList[len(previousRequestList)-1].URL.String(), "to", nextRequest.URL.String()) // log each redirect hop
			if len(previousRequestList) >= maximumRedirectCount {                                                                                                                         // check whether there were too many redirects
				return fmt.Errorf("stopped after %d redirects", maximumRedirectCount) // give up
			} // end of redirect limit check
			return nil // allow the redirect
		}, // end of redirect function
	} // end of client creation
	slog.Debug("http client created", "timeout", requestTimeout) // confirm the client is ready

	alreadySeenLinkSet := make(map[string]bool) // remembers links already handled in this run
	downloadedCount := 0                        // counts newly downloaded files
	skippedCount := 0                           // counts files skipped because they already exist
	failedCount := 0                            // counts files that failed
	totalBytesDownloaded := int64(0)            // counts all bytes written to disk in this run

	for pageNumber := firstPageNumber; pageNumber <= lastPageNumber; pageNumber++ { // visit page 1 through page 10
		pageStartTime := time.Now()                                          // remember when this page started
		pageUrl := buildPageUrl(pageNumber)                                  // create the url for this page number
		slog.Info("fetching page", "pageNumber", pageNumber, "url", pageUrl) // show which page is being fetched

		pageContent, fetchPageError := fetchPageContent(httpClient, pageUrl) // download the page HTML
		if fetchPageError != nil {                                           // check whether the page could not be fetched
			slog.Error("failed to fetch page", "pageNumber", pageNumber, "url", pageUrl, "error", fetchPageError) // log the problem
			continue                                                                                              // move on to the next page number
		} // end of page error check
		slog.Debug("page fetched", "pageNumber", pageNumber, "contentLengthCharacters", len(pageContent), "elapsed", time.Since(pageStartTime)) // show page size and time

		portableDocumentFileLinkList := extractPortableDocumentFileLinks(pageContent, pageUrl)                 // find all PDF links on the page
		slog.Info("PDF links found", "pageNumber", pageNumber, "linkCount", len(portableDocumentFileLinkList)) // show how many were found

		for linkIndex, portableDocumentFileLink := range portableDocumentFileLinkList { // go through each PDF link
			slog.Debug("processing link", "pageNumber", pageNumber, "linkIndex", linkIndex+1, "linkTotal", len(portableDocumentFileLinkList), "link", portableDocumentFileLink) // show which link is being handled

			if alreadySeenLinkSet[portableDocumentFileLink] { // check whether this link was already handled
				slog.Debug("duplicate link seen on an earlier page, skipping", "link", portableDocumentFileLink) // explain the skip
				continue                                                                                         // skip duplicates
			} // end of duplicate check
			alreadySeenLinkSet[portableDocumentFileLink] = true // mark this link as handled

			fileName, fileNameError := buildFileNameFromLink(portableDocumentFileLink) // work out the local file name
			if fileNameError != nil {                                                  // check whether a name could not be built
				slog.Error("could not build file name", "link", portableDocumentFileLink, "error", fileNameError) // log the problem
				failedCount++                                                                                     // count it as a failure
				continue                                                                                          // move on to the next link
			} // end of file name error check

			destinationPath := filepath.Join(outputDirectoryName, fileName)                             // full local path such as PDFs/example.pdf
			slog.Debug("destination decided", "fileName", fileName, "destinationPath", destinationPath) // show where it will be saved

			existingFileInformation, statError := os.Stat(destinationPath) // check whether the file is already on disk
			if statError == nil {                                          // no error means the file exists
				slog.Info("already downloaded, skipping", "fileName", fileName, "existingSizeBytes", existingFileInformation.Size(), "modifiedTime", existingFileInformation.ModTime()) // tell the user it is being skipped
				skippedCount++                                                                                                                                                          // count the skip
				continue                                                                                                                                                                // move on to the next link
			} // end of existing file check
			if !os.IsNotExist(statError) { // any error other than "file does not exist" is unexpected
				slog.Error("could not check destination file", "destinationPath", destinationPath, "error", statError) // log the unexpected error
				failedCount++                                                                                          // count it as a failure
				continue                                                                                               // move on to the next link
			} // end of unexpected error check
			slog.Debug("file not on disk yet, will download", "destinationPath", destinationPath) // confirm a download is needed

			slog.Info("downloading", "link", portableDocumentFileLink)                                               // show which file is being downloaded
			downloadStartTime := time.Now()                                                                          // remember when this download started
			bytesWritten, downloadError := downloadFileToPath(httpClient, portableDocumentFileLink, destinationPath) // download and save the file
			if downloadError != nil {                                                                                // check whether the download failed
				slog.Error("download failed", "link", portableDocumentFileLink, "error", downloadError, "elapsed", time.Since(downloadStartTime)) // log the problem
				failedCount++                                                                                                                     // count it as a failure
				continue                                                                                                                          // move on to the next link
			} // end of download error check

			downloadDuration := time.Since(downloadStartTime) // measure how long the download took
			bytesPerSecond := float64(0)                      // download speed, starts at zero
			if downloadDuration.Seconds() > 0 {               // avoid dividing by zero
				bytesPerSecond = float64(bytesWritten) / downloadDuration.Seconds() // calculate speed
			} // end of speed calculation
			slog.Info("saved", "destinationPath", destinationPath, "sizeBytes", bytesWritten, "elapsed", downloadDuration, "bytesPerSecond", int64(bytesPerSecond)) // confirm the file was saved
			downloadedCount++                                                                                                                                       // count the successful download
			totalBytesDownloaded += bytesWritten                                                                                                                    // add to the running total

			slog.Debug("sleeping before next request", "delay", delayBetweenRequests) // announce the pause
			time.Sleep(delayBetweenRequests)                                          // wait briefly before the next request
		} // end of loop over PDF links

		slog.Debug("finished page", "pageNumber", pageNumber, "elapsed", time.Since(pageStartTime)) // show how long the whole page took
	} // end of loop over page numbers

	slog.Info("done", "downloaded", downloadedCount, "skipped", skippedCount, "failed", failedCount, "totalBytesDownloaded", totalBytesDownloaded, "uniqueLinksSeen", len(alreadySeenLinkSet), "totalElapsed", time.Since(programStartTime)) // print the final summary
} // end of main

func buildPageUrl(pageNumber int) string { // returns the url for a given page number
	if pageNumber == 1 { // page 1 has no /page/N/ part
		slog.Debug("building url for first page", "pageNumber", pageNumber) // explain the special case
		return downloadPageBaseUrl                                          // return the base downloads url
	} // end of page 1 check
	return fmt.Sprintf("%spage/%d/", downloadPageBaseUrl, pageNumber) // pages 2 and up look like /downloads/page/2/
} // end of buildPageUrl

func buildClientTrace(requestUrl string) *httptrace.ClientTrace { // creates hooks that log low-level network events
	traceStartTime := time.Now()   // remember when the request began so each event can show elapsed time
	return &httptrace.ClientTrace{ // return the set of hooks
		GetConn: func(hostPort string) { // called when a connection is requested
			slog.Debug("trace: getting connection", "hostPort", hostPort, "elapsed", time.Since(traceStartTime)) // log the request for a connection
		}, // end of GetConn
		GotConn: func(gotConnectionInformation httptrace.GotConnInfo) { // called when a connection is ready
			slog.Debug("trace: got connection", "remoteAddress", gotConnectionInformation.Conn.RemoteAddr().String(), "reused", gotConnectionInformation.Reused, "wasIdle", gotConnectionInformation.WasIdle, "idleTime", gotConnectionInformation.IdleTime, "elapsed", time.Since(traceStartTime)) // log connection details
		}, // end of GotConn
		DNSStart: func(dnsStartInformation httptrace.DNSStartInfo) { // called when a DNS lookup begins
			slog.Debug("trace: DNS lookup started", "host", dnsStartInformation.Host, "elapsed", time.Since(traceStartTime)) // log the host being looked up
		}, // end of DNSStart
		DNSDone: func(dnsDoneInformation httptrace.DNSDoneInfo) { // called when a DNS lookup ends
			slog.Debug("trace: DNS lookup done", "addresses", fmt.Sprint(dnsDoneInformation.Addrs), "coalesced", dnsDoneInformation.Coalesced, "error", dnsDoneInformation.Err, "elapsed", time.Since(traceStartTime)) // log the addresses found
		}, // end of DNSDone
		ConnectStart: func(network string, address string) { // called when a TCP connection attempt begins
			slog.Debug("trace: connecting", "network", network, "address", address, "elapsed", time.Since(traceStartTime)) // log the attempt
		}, // end of ConnectStart
		ConnectDone: func(network string, address string, connectError error) { // called when a TCP connection attempt ends
			slog.Debug("trace: connect done", "network", network, "address", address, "error", connectError, "elapsed", time.Since(traceStartTime)) // log the result
		}, // end of ConnectDone
		TLSHandshakeStart: func() { // called when the TLS handshake begins
			slog.Debug("trace: TLS handshake started", "elapsed", time.Since(traceStartTime)) // log the start
		}, // end of TLSHandshakeStart
		TLSHandshakeDone: func(connectionState tls.ConnectionState, handshakeError error) { // called when the TLS handshake ends
			slog.Debug("trace: TLS handshake done", "tlsVersion", tls.VersionName(connectionState.Version), "cipherSuite", tls.CipherSuiteName(connectionState.CipherSuite), "negotiatedProtocol", connectionState.NegotiatedProtocol, "serverName", connectionState.ServerName, "error", handshakeError, "elapsed", time.Since(traceStartTime)) // log the TLS details
		}, // end of TLSHandshakeDone
		WroteRequest: func(wroteRequestInformation httptrace.WroteRequestInfo) { // called when the request has been sent
			slog.Debug("trace: request written", "url", requestUrl, "error", wroteRequestInformation.Err, "elapsed", time.Since(traceStartTime)) // log that the request was sent
		}, // end of WroteRequest
		GotFirstResponseByte: func() { // called when the first byte of the response arrives
			slog.Debug("trace: first response byte received", "url", requestUrl, "elapsed", time.Since(traceStartTime)) // log time to first byte
		}, // end of GotFirstResponseByte
	} // end of trace hooks
} // end of buildClientTrace

func sendGetRequest(httpClient *http.Client, requestUrl string) (*http.Response, error) { // sends a GET request and checks the status
	slog.Debug("creating request", "method", http.MethodGet, "url", requestUrl)      // announce the request
	httpRequest, newRequestError := http.NewRequest(http.MethodGet, requestUrl, nil) // create the request object
	if newRequestError != nil {                                                      // check whether the request could not be created
		return nil, newRequestError // give the error to the caller
	} // end of request creation check
	httpRequest.Header.Set("User-Agent", browserUserAgent)                  // identify as a normal browser
	httpRequest.Header.Set("Accept", "text/html,application/pdf,*/*;q=0.8") // say which content types are acceptable

	tracedContext := httptrace.WithClientTrace(httpRequest.Context(), buildClientTrace(requestUrl)) // attach the network trace hooks
	httpRequest = httpRequest.WithContext(tracedContext)                                            // use the traced context for this request

	for requestHeaderName, requestHeaderValueList := range httpRequest.Header { // go through every request header
		slog.Debug("request header", "name", requestHeaderName, "values", strings.Join(requestHeaderValueList, ", ")) // log each header
	} // end of request header loop

	requestStartTime := time.Now()                             // remember when the request was sent
	httpResponse, doRequestError := httpClient.Do(httpRequest) // send the request
	if doRequestError != nil {                                 // check whether the request failed
		slog.Debug("request failed", "url", requestUrl, "error", doRequestError, "elapsed", time.Since(requestStartTime)) // log the failure
		return nil, doRequestError                                                                                        // give the error to the caller
	} // end of request error check

	slog.Debug("response received", "url", requestUrl, "finalUrl", httpResponse.Request.URL.String(), "status", httpResponse.Status, "protocol", httpResponse.Proto, "contentLength", httpResponse.ContentLength, "contentType", httpResponse.Header.Get("Content-Type"), "elapsed", time.Since(requestStartTime)) // log the key response facts
	for responseHeaderName, responseHeaderValueList := range httpResponse.Header {                                                                                                                                                                                                                                 // go through every response header
		slog.Debug("response header", "name", responseHeaderName, "values", strings.Join(responseHeaderValueList, ", ")) // log each header
	} // end of response header loop

	if httpResponse.StatusCode != http.StatusOK { // anything other than 200 OK is treated as a failure
		httpResponse.Body.Close()                                            // close the body to avoid leaking the connection
		return nil, fmt.Errorf("unexpected status: %s", httpResponse.Status) // report the status, such as 404 Not Found
	} // end of status check
	return httpResponse, nil // return the successful response
} // end of sendGetRequest

func fetchPageContent(httpClient *http.Client, pageUrl string) (string, error) { // downloads a page and returns its HTML text
	httpResponse, sendRequestError := sendGetRequest(httpClient, pageUrl) // request the page
	if sendRequestError != nil {                                          // check whether the request failed
		return "", sendRequestError // give the error to the caller
	} // end of request error check
	defer httpResponse.Body.Close() // close the body when this function ends

	slog.Debug("reading page body", "url", pageUrl)                   // announce the read
	responseBodyBytes, readBodyError := io.ReadAll(httpResponse.Body) // read the whole page into memory
	if readBodyError != nil {                                         // check whether reading failed
		return "", readBodyError // give the error to the caller
	} // end of read error check
	slog.Debug("page body read", "url", pageUrl, "bytesRead", len(responseBodyBytes)) // show how many bytes were read
	return string(responseBodyBytes), nil                                             // convert the bytes to text and return it
} // end of fetchPageContent

func extractPortableDocumentFileLinks(pageContent string, pageUrl string) []string { // finds all unique PDF links in the HTML
	baseUrl, parseBaseUrlError := url.Parse(pageUrl) // parse the page url so relative links can be resolved
	if parseBaseUrlError != nil {                    // check whether parsing failed
		slog.Error("could not parse base url", "url", pageUrl, "error", parseBaseUrlError) // log the problem
		return nil                                                                         // return no links
	} // end of base url check

	matchGroupListList := portableDocumentFileLinkPattern.FindAllStringSubmatch(pageContent, -1)                                      // find every regex match in the page
	slog.Debug("regex matching finished", "pattern", portableDocumentFileLinkPattern.String(), "matchCount", len(matchGroupListList)) // show how many raw matches there were

	uniqueLinkSet := make(map[string]bool) // remembers links already added to the list
	var linkList []string                  // the list of links that will be returned

	for matchIndex, matchGroupList := range matchGroupListList { // loop over every regex match
		rawLink := html.UnescapeString(strings.TrimSpace(matchGroupList[1]))                                      // take the attribute value, trim spaces, decode entities
		slog.Debug("regex match", "matchIndex", matchIndex+1, "fullMatch", matchGroupList[0], "rawLink", rawLink) // show exactly what was matched

		resolvedUrl, resolveError := baseUrl.Parse(rawLink) // turn a relative link into a full url
		if resolveError != nil {                            // check whether resolving failed
			slog.Warn("could not resolve link", "rawLink", rawLink, "error", resolveError) // log the problem
			continue                                                                       // move on to the next match
		} // end of resolve error check

		resolvedLink := resolvedUrl.String() // convert the url back to text
		if uniqueLinkSet[resolvedLink] {     // check whether it is already in the list
			slog.Debug("duplicate link on the same page, ignoring", "link", resolvedLink) // explain the skip
			continue                                                                      // skip duplicates
		} // end of duplicate check
		uniqueLinkSet[resolvedLink] = true                // remember this link
		linkList = append(linkList, resolvedLink)         // add it to the result list
		slog.Debug("link accepted", "link", resolvedLink) // confirm the link was kept
	} // end of loop over matches

	return linkList // return all unique PDF links
} // end of extractPortableDocumentFileLinks

func buildFileNameFromLink(portableDocumentFileLink string) (string, error) { // turns a PDF url into a safe local file name
	parsedUrl, parseUrlError := url.Parse(portableDocumentFileLink) // parse the url
	if parseUrlError != nil {                                       // check whether parsing failed
		return "", parseUrlError // give the error to the caller
	} // end of parse error check

	fileName := path.Base(parsedUrl.Path)                                                  // take the last part of the url path
	slog.Debug("file name from url path", "urlPath", parsedUrl.Path, "baseName", fileName) // show the raw name
	if fileName == "." || fileName == "/" || fileName == "" {                              // check for an unusable name
		return "", fmt.Errorf("could not determine file name from %q", portableDocumentFileLink) // report the problem
	} // end of unusable name check

	sanitizedFileName := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(fileName) // replace characters that are unsafe in file names
	if sanitizedFileName != fileName {                                                        // check whether anything was changed
		slog.Debug("file name sanitized", "before", fileName, "after", sanitizedFileName) // show the change
	} // end of sanitize check
	fileName = sanitizedFileName // use the sanitized name from here on

	if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") { // make sure the name ends with .pdf
		slog.Debug("adding missing .pdf extension", "fileName", fileName) // explain the change
		fileName += ".pdf"                                                // add the extension if it is missing
	} // end of extension check
	return fileName, nil // return the final file name
} // end of buildFileNameFromLink

func downloadFileToPath(httpClient *http.Client, fileUrl string, destinationPath string) (int64, error) { // downloads a file, saves it, and returns the number of bytes written
	httpResponse, sendRequestError := sendGetRequest(httpClient, fileUrl) // request the file
	if sendRequestError != nil {                                          // check whether the request failed
		return 0, sendRequestError // give the error to the caller
	} // end of request error check
	defer httpResponse.Body.Close() // close the body when this function ends

	temporaryPath := destinationPath + ".part"                            // save to a temporary name first so partial downloads are never mistaken for finished ones
	slog.Debug("creating temporary file", "temporaryPath", temporaryPath) // announce the file creation
	temporaryFile, createFileError := os.Create(temporaryPath)            // create the temporary file
	if createFileError != nil {                                           // check whether creating the file failed
		return 0, createFileError // give the error to the caller
	} // end of create error check

	slog.Debug("copying response body to disk", "expectedBytes", httpResponse.ContentLength)                    // announce the copy
	bytesWritten, copyError := io.Copy(temporaryFile, httpResponse.Body)                                        // stream the response body into the file
	closeError := temporaryFile.Close()                                                                         // close the file so everything is written to disk
	slog.Debug("copy finished", "bytesWritten", bytesWritten, "copyError", copyError, "closeError", closeError) // show the outcome

	if copyError != nil { // check whether copying failed
		slog.Debug("removing partial file after copy error", "temporaryPath", temporaryPath) // explain the cleanup
		os.Remove(temporaryPath)                                                             // delete the partial file
		return bytesWritten, copyError                                                       // give the error to the caller
	} // end of copy error check
	if closeError != nil { // check whether closing failed
		slog.Debug("removing partial file after close error", "temporaryPath", temporaryPath) // explain the cleanup
		os.Remove(temporaryPath)                                                              // delete the partial file
		return bytesWritten, closeError                                                       // give the error to the caller
	} // end of close error check
	if httpResponse.ContentLength >= 0 && bytesWritten != httpResponse.ContentLength { // compare the size on disk with what the server promised
		slog.Warn("size differs from Content-Length header", "bytesWritten", bytesWritten, "contentLength", httpResponse.ContentLength) // warn about a mismatch
	} // end of size check

	slog.Debug("renaming temporary file", "from", temporaryPath, "to", destinationPath) // announce the rename
	renameError := os.Rename(temporaryPath, destinationPath)                            // rename the finished file to its final name
	if renameError != nil {                                                             // check whether renaming failed
		os.Remove(temporaryPath)         // delete the temporary file
		return bytesWritten, renameError // give the error to the caller
	} // end of rename error check
	return bytesWritten, nil // success
} // end of downloadFileToPath
