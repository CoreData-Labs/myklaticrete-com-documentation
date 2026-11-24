package main // this is an executable program

import ( // standard library packages used by this program
	"fmt"           // building formatted error messages and urls
	"html"          // decoding HTML entities such as &amp; inside links
	"io"            // reading response bodies and copying them to files
	"log"           // printing progress and error messages
	"net/http"      // making HTTP requests
	"net/url"       // parsing and resolving urls
	"os"            // creating folders and files, checking if files exist
	"path"          // getting the last element of a url path
	"path/filepath" // building local file paths for the current operating system
	"regexp"        // finding PDF links inside the page HTML
	"strings"       // trimming and replacing text
	"time"          // timeouts and delays
) // end of imports

const ( // fixed settings for the program
	outputDirectoryName  = "PDFs"                                                                                                        // folder where the PDFs are saved
	firstPageNumber      = 1                                                                                                             // first download page to visit
	lastPageNumber       = 10                                                                                                            // last download page to visit
	downloadPageBaseUrl  = "https://myklaticrete.com/downloads/"                                                                         // url of page 1
	requestTimeout       = 60 * time.Second                                                                                              // maximum time for one HTTP request
	delayBetweenRequests = 500 * time.Millisecond                                                                                        // pause between downloads to be polite to the server
	browserUserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36" // pretend to be a normal browser
) // end of constants

var portableDocumentFileLinkPattern = regexp.MustCompile(`(?i)(?:data-pdf-url|href)\s*=\s*["']([^"']+?\.pdf(?:\?[^"']*)?)["']`) // matches data-pdf-url="...pdf" (and href="...pdf") attributes, ignoring letter case

func main() { // program starts here
	log.SetFlags(log.LstdFlags | log.Lshortfile) // log the date, time, and source line with every message

	makeDirectoryError := os.MkdirAll(outputDirectoryName, 0o755) // create the PDFs folder if it does not exist
	if makeDirectoryError != nil {                                // check whether the folder could not be created
		log.Fatalf("could not create output directory %q: %v", outputDirectoryName, makeDirectoryError) // print the error and stop
	} // end of folder error check

	httpClient := &http.Client{Timeout: requestTimeout} // one reusable HTTP client with a timeout

	alreadySeenLinkSet := make(map[string]bool) // remembers links already handled in this run
	downloadedCount := 0                        // counts newly downloaded files
	skippedCount := 0                           // counts files skipped because they already exist
	failedCount := 0                            // counts files that failed

	for pageNumber := firstPageNumber; pageNumber <= lastPageNumber; pageNumber++ { // visit page 1 through page 10
		pageUrl := buildPageUrl(pageNumber)                     // create the url for this page number
		log.Printf("fetching page %d: %s", pageNumber, pageUrl) // show which page is being fetched

		pageContent, fetchPageError := fetchPageContent(httpClient, pageUrl) // download the page HTML
		if fetchPageError != nil {                                           // check whether the page could not be fetched
			log.Printf("failed to fetch page %s: %v", pageUrl, fetchPageError) // log the problem
			continue                                                           // move on to the next page number
		} // end of page error check

		portableDocumentFileLinkList := extractPortableDocumentFileLinks(pageContent, pageUrl)       // find all PDF links on the page
		log.Printf("found %d PDF link(s) on page %d", len(portableDocumentFileLinkList), pageNumber) // show how many were found

		for _, portableDocumentFileLink := range portableDocumentFileLinkList { // go through each PDF link
			if alreadySeenLinkSet[portableDocumentFileLink] { // check whether this link was already handled
				continue // skip duplicates found on multiple pages
			} // end of duplicate check
			alreadySeenLinkSet[portableDocumentFileLink] = true // mark this link as handled

			fileName, fileNameError := buildFileNameFromLink(portableDocumentFileLink) // work out the local file name
			if fileNameError != nil {                                                  // check whether a name could not be built
				log.Printf("skipping %s: %v", portableDocumentFileLink, fileNameError) // log the problem
				failedCount++                                                          // count it as a failure
				continue                                                               // move on to the next link
			} // end of file name error check

			destinationPath := filepath.Join(outputDirectoryName, fileName) // full local path such as PDFs/example.pdf

			_, statError := os.Stat(destinationPath) // check whether the file is already on disk
			if statError == nil {                    // no error means the file exists
				log.Printf("already downloaded, skipping: %s", fileName) // tell the user it is being skipped
				skippedCount++                                           // count the skip
				continue                                                 // move on to the next link
			} // end of existing file check
			if !os.IsNotExist(statError) { // any error other than "file does not exist" is unexpected
				log.Printf("could not check %s: %v", destinationPath, statError) // log the unexpected error
				failedCount++                                                    // count it as a failure
				continue                                                         // move on to the next link
			} // end of unexpected error check

			log.Printf("downloading: %s", portableDocumentFileLink)                                    // show which file is being downloaded
			downloadError := downloadFileToPath(httpClient, portableDocumentFileLink, destinationPath) // download and save the file
			if downloadError != nil {                                                                  // check whether the download failed
				log.Printf("failed to download %s: %v", portableDocumentFileLink, downloadError) // log the problem
				failedCount++                                                                    // count it as a failure
				continue                                                                         // move on to the next link
			} // end of download error check

			log.Printf("saved: %s", destinationPath) // confirm the file was saved
			downloadedCount++                        // count the successful download
			time.Sleep(delayBetweenRequests)         // wait briefly before the next request
		} // end of loop over PDF links
	} // end of loop over page numbers

	log.Printf("done. downloaded=%d skipped=%d failed=%d", downloadedCount, skippedCount, failedCount) // print the final summary
} // end of main

func buildPageUrl(pageNumber int) string { // returns the url for a given page number
	if pageNumber == 1 { // page 1 has no /page/N/ part
		return downloadPageBaseUrl // return the base downloads url
	} // end of page 1 check
	return fmt.Sprintf("%spage/%d/", downloadPageBaseUrl, pageNumber) // pages 2 and up look like /downloads/page/2/
} // end of buildPageUrl

func sendGetRequest(httpClient *http.Client, requestUrl string) (*http.Response, error) { // sends a GET request and checks the status
	httpRequest, newRequestError := http.NewRequest(http.MethodGet, requestUrl, nil) // create the request object
	if newRequestError != nil {                                                      // check whether the request could not be created
		return nil, newRequestError // give the error to the caller
	} // end of request creation check
	httpRequest.Header.Set("User-Agent", browserUserAgent)                  // identify as a normal browser
	httpRequest.Header.Set("Accept", "text/html,application/pdf,*/*;q=0.8") // say which content types are acceptable

	httpResponse, doRequestError := httpClient.Do(httpRequest) // send the request
	if doRequestError != nil {                                 // check whether the request failed
		return nil, doRequestError // give the error to the caller
	} // end of request error check
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

	responseBodyBytes, readBodyError := io.ReadAll(httpResponse.Body) // read the whole page into memory
	if readBodyError != nil {                                         // check whether reading failed
		return "", readBodyError // give the error to the caller
	} // end of read error check
	return string(responseBodyBytes), nil // convert the bytes to text and return it
} // end of fetchPageContent

func extractPortableDocumentFileLinks(pageContent string, pageUrl string) []string { // finds all unique PDF links in the HTML
	baseUrl, parseBaseUrlError := url.Parse(pageUrl) // parse the page url so relative links can be resolved
	if parseBaseUrlError != nil {                    // check whether parsing failed
		log.Printf("could not parse base url %s: %v", pageUrl, parseBaseUrlError) // log the problem
		return nil                                                                // return no links
	} // end of base url check

	uniqueLinkSet := make(map[string]bool) // remembers links already added to the list
	var linkList []string                  // the list of links that will be returned

	for _, matchGroupList := range portableDocumentFileLinkPattern.FindAllStringSubmatch(pageContent, -1) { // loop over every regex match
		rawLink := html.UnescapeString(strings.TrimSpace(matchGroupList[1])) // take the href value, trim spaces, decode entities

		resolvedUrl, resolveError := baseUrl.Parse(rawLink) // turn a relative link into a full url
		if resolveError != nil {                            // check whether resolving failed
			log.Printf("could not resolve link %q: %v", rawLink, resolveError) // log the problem
			continue                                                           // move on to the next match
		} // end of resolve error check

		resolvedLink := resolvedUrl.String() // convert the url back to text
		if uniqueLinkSet[resolvedLink] {     // check whether it is already in the list
			continue // skip duplicates
		} // end of duplicate check
		uniqueLinkSet[resolvedLink] = true        // remember this link
		linkList = append(linkList, resolvedLink) // add it to the result list
	} // end of loop over matches

	return linkList // return all unique PDF links
} // end of extractPortableDocumentFileLinks

func buildFileNameFromLink(portableDocumentFileLink string) (string, error) { // turns a PDF url into a safe local file name
	parsedUrl, parseUrlError := url.Parse(portableDocumentFileLink) // parse the url
	if parseUrlError != nil {                                       // check whether parsing failed
		return "", parseUrlError // give the error to the caller
	} // end of parse error check

	fileName := path.Base(parsedUrl.Path)                     // take the last part of the url path
	if fileName == "." || fileName == "/" || fileName == "" { // check for an unusable name
		return "", fmt.Errorf("could not determine file name from %q", portableDocumentFileLink) // report the problem
	} // end of unusable name check

	fileName = strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(fileName) // replace characters that are unsafe in file names

	if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") { // make sure the name ends with .pdf
		fileName += ".pdf" // add the extension if it is missing
	} // end of extension check
	return fileName, nil // return the final file name
} // end of buildFileNameFromLink

func downloadFileToPath(httpClient *http.Client, fileUrl string, destinationPath string) error { // downloads a file and saves it
	httpResponse, sendRequestError := sendGetRequest(httpClient, fileUrl) // request the file
	if sendRequestError != nil {                                          // check whether the request failed
		return sendRequestError // give the error to the caller
	} // end of request error check
	defer httpResponse.Body.Close() // close the body when this function ends

	temporaryPath := destinationPath + ".part"                 // save to a temporary name first so partial downloads are never mistaken for finished ones
	temporaryFile, createFileError := os.Create(temporaryPath) // create the temporary file
	if createFileError != nil {                                // check whether creating the file failed
		return createFileError // give the error to the caller
	} // end of create error check

	_, copyError := io.Copy(temporaryFile, httpResponse.Body) // stream the response body into the file
	closeError := temporaryFile.Close()                       // close the file so everything is written to disk

	if copyError != nil { // check whether copying failed
		os.Remove(temporaryPath) // delete the partial file
		return copyError         // give the error to the caller
	} // end of copy error check
	if closeError != nil { // check whether closing failed
		os.Remove(temporaryPath) // delete the partial file
		return closeError        // give the error to the caller
	} // end of close error check

	renameError := os.Rename(temporaryPath, destinationPath) // rename the finished file to its final name
	if renameError != nil {                                  // check whether renaming failed
		os.Remove(temporaryPath) // delete the temporary file
		return renameError       // give the error to the caller
	} // end of rename error check
	return nil // success
} // end of downloadFileToPath
