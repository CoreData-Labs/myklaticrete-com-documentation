# 📚 MyKlaticrete Documentation Archive

> A bulk archive of the technical documents published on [myklaticrete.com/downloads](https://myklaticrete.com/downloads/), collected automatically with a dependency-free **Go** downloader. 🏗️🧪📄

![Go](https://img.shields.io/badge/Go-standard%20library%20only-00ADD8?logo=go&logoColor=white)
![Format](https://img.shields.io/badge/format-PDF-red)
![Topics](https://img.shields.io/badge/topics-SDS%20%7C%20MSDS%20%7C%20TDS%20%7C%20Install%20Guides-blue)
![Status](https://img.shields.io/badge/status-archival%20%2F%20educational-lightgrey)

---

## 📑 Table of Contents

1. [🎯 What Is This?](#-what-is-this)
2. [📦 What's Inside](#-whats-inside)
3. [🗂️ Repository Structure](#️-repository-structure)
4. [🤖 How the Data Was Collected](#-how-the-data-was-collected)
5. [🚀 Quick Start](#-quick-start)
6. [🧾 Understanding an SDS / MSDS](#-understanding-an-sds--msds)
7. [🔍 How to Use This Data](#-how-to-use-this-data)
8. [🛡️ Benefit: Improving Safety](#️-benefit-improving-safety)
9. [✅ Benefit: Improving Quality](#-benefit-improving-quality)
10. [🔬 Benefit: Reverse Engineering & Product Analysis](#-benefit-reverse-engineering--product-analysis)
11. [📊 More Use Cases](#-more-use-cases)
12. [⚠️ Limitations](#️-limitations)
13. [⚖️ Legal & Disclaimer](#️-legal--disclaimer)
14. [🤝 Contributing](#-contributing)
15. [👤 Author](#-author)

---

## 🎯 What Is This?

Construction product manufacturers publish a large amount of technical paperwork: safety data sheets, technical data sheets, installation guides, certifications, warranties and more. That information is valuable, but it is scattered across many web pages and is hard to search, compare or keep offline.

This repository gathers those PDFs into one place so that you can:

- 📥 **Keep an offline copy** of the documents for job sites with poor internet
- 🔎 **Search across every document at once** instead of opening them one by one
- 📊 **Extract and compare data** (ingredients, hazards, cure times, standards)
- 🧠 **Study how products are described, classified and certified**
- 🛡️ **Build safety and compliance workflows** on top of real documents

---

## 📦 What's Inside

The archive is built from the PDF links found on the download pages (pages 1 to 10). Document types you can expect to find:

| Icon | Document Type                          | What It Tells You                                                                                  |
| ---- | -------------------------------------- | -------------------------------------------------------------------------------------------------- |
| 🧪   | **SDS / MSDS** (Safety Data Sheets)    | Hazards, ingredients, first aid, fire-fighting, spill response, storage, disposal, regulatory info |
| 📐   | **TDS** (Technical Data Sheets)        | Product description, coverage, mixing ratios, pot life, cure times, performance values             |
| 🛠️   | **Installation Guides / Instructions** | Surface preparation, application steps, tools, curing, maintenance                                 |
| 🏅   | **Certifications & Test Reports**      | Standards compliance (e.g. ANSI, ISO, ASTM references) and test results                            |
| 📜   | **Warranties**                         | Coverage terms, conditions and exclusions                                                          |
| 📘   | **Brochures & Specification Sheets**   | System overviews, product selection guidance, specification text                                   |

> 📝 **Note:** The exact mix of documents depends on what the website published at the time the archive was built. Run the commands in [Quick Start](#-quick-start) to see the real counts for your copy.

---

## 🗂️ Repository Structure

```text
myklaticrete-com-documentation/
├── 📄 README.md        # you are here
├── 🧾 main.go          # Go downloader (standard library only)
└── 📁 PDFs/            # all downloaded PDF documents
    ├── Clenza-TS_MSDS.pdf
    └── ...
```

---

## 🤖 How the Data Was Collected

The files were gathered with `main.go`, a small downloader written using **only the Go standard library** (`net/http`, `regexp`, `net/url`, `os`, `io`, `log` and friends).

**How it works:**

1. 🔁 Loops over download pages **1 → 10**
   - Page 1: `https://myklaticrete.com/downloads/`
   - Pages 2+: `https://myklaticrete.com/downloads/page/N/`
2. 🔗 Finds every PDF link. On this site the link sits in a `data-pdf-url="..."` attribute, so the scraper reads that (and plain `href` links too)
3. ♻️ **Skips files you already have** in `PDFs/`
4. 💾 Downloads to a temporary `.part` file, then renames it, so half-finished downloads are never mistaken for complete ones
5. 📝 Logs progress and prints a summary (`downloaded / skipped / failed`)

### 🔄 Updating the archive

Just run the program again. Only new files are downloaded:

```bash
go run main.go
```

---

## 🚀 Quick Start

### 1️⃣ Clone the repository

```bash
git clone https://github.com/CoreData-Labs/myklaticrete-com-documentation.git
cd myklaticrete-com-documentation
```

### 2️⃣ Check what you have

```bash
# number of PDFs
ls PDFs/*.pdf | wc -l

# total size on disk
du -sh PDFs/

# all safety data sheets (file names containing SDS or MSDS)
ls PDFs/ | grep -i -E "sds|msds"
```

### 3️⃣ (Optional) Refresh the archive

```bash
# requires Go 1.20+
go run main.go
```

### 4️⃣ Install a PDF text tool for searching

```bash
# macOS
brew install poppler

# Ubuntu / Debian
sudo apt install poppler-utils

# Windows: install poppler from https://github.com/oschwartz10612/poppler-windows
```

---

## 🧾 Understanding an SDS / MSDS

**MSDS** (Material Safety Data Sheet) is the older name. **SDS** (Safety Data Sheet) is the current name under the **GHS** (Globally Harmonized System), which standardizes SDS layouts into **16 sections**:

| #   | Section                                  | Why You Care                                                            |
| --- | ---------------------------------------- | ----------------------------------------------------------------------- |
| 1   | Identification                           | Product name, supplier, emergency phone number                          |
| 2   | Hazard(s) identification                 | Hazard classes, signal word, pictograms, precautionary statements ⚠️    |
| 3   | Composition / information on ingredients | Hazardous ingredients, **CAS numbers**, concentration ranges 🧪         |
| 4   | First-aid measures                       | What to do after inhalation, skin/eye contact, ingestion 🚑             |
| 5   | Fire-fighting measures                   | Suitable extinguishing media, special hazards 🔥                        |
| 6   | Accidental release measures              | Spill containment and clean-up 🧹                                       |
| 7   | Handling and storage                     | Safe handling practices, storage conditions, incompatibilities 📦       |
| 8   | Exposure controls / personal protection  | Exposure limits, ventilation, gloves, eye and respiratory protection 🥽 |
| 9   | Physical and chemical properties         | Appearance, pH, density, flash point, solubility 🌡️                     |
| 10  | Stability and reactivity                 | Conditions and materials to avoid, hazardous reactions                  |
| 11  | Toxicological information                | Health effects, routes of exposure                                      |
| 12  | Ecological information                   | Environmental impact 🌍                                                 |
| 13  | Disposal considerations                  | Waste handling guidance ♻️                                              |
| 14  | Transport information                    | UN number, shipping class                                               |
| 15  | Regulatory information                   | Applicable regulations and listings                                     |
| 16  | Other information                        | Revision date, abbreviations, references 📅                             |

> 💡 **Tip:** Always check **Section 16** for the revision date. A data sheet is only as trustworthy as it is current.

---

## 🔍 How to Use This Data

### 🔎 Search every PDF for a keyword

```bash
# convert all PDFs to text once
mkdir -p TXT
for file in PDFs/*.pdf; do
  pdftotext -layout "$file" "TXT/$(basename "${file%.pdf}").txt"
done

# search all of them
grep -ril "silica" TXT/
```

### 🧬 Extract every CAS number across the archive

CAS numbers identify chemical substances (format `1234-56-7`).

```bash
grep -ohE "\b[0-9]{2,7}-[0-9]{2}-[0-9]\b" TXT/*.txt | sort | uniq -c | sort -rn | head -30
```

### 📋 Find which products contain a specific ingredient

```bash
# e.g. everything that lists crystalline silica (CAS 14808-60-7)
grep -l "14808-60-7" TXT/*.txt
```

### 🔥 Pull out flash points, pH and VOC values

```bash
grep -iE "flash point|pH|VOC" TXT/*.txt
```

### 🐍 Build a spreadsheet-friendly index (Python example)

```python
import csv, re, subprocess
from pathlib import Path

CAS_PATTERN = re.compile(r"\b\d{2,7}-\d{2}-\d\b")
rows = []

for pdf_path in sorted(Path("PDFs").glob("*.pdf")):
    text = subprocess.run(
        ["pdftotext", "-layout", str(pdf_path), "-"],
        capture_output=True, text=True,
    ).stdout
    rows.append({
        "file": pdf_path.name,
        "is_sds": bool(re.search(r"safety data sheet|msds", text, re.I)),
        "cas_numbers": ";".join(sorted(set(CAS_PATTERN.findall(text)))),
        "pages_text_chars": len(text),
    })

with open("document_index.csv", "w", newline="") as output_file:
    writer = csv.DictWriter(output_file, fieldnames=rows[0].keys())
    writer.writeheader()
    writer.writerows(rows)
```

Open `document_index.csv` in Excel or Google Sheets to filter, sort and pivot. 📈

---

## 🛡️ Benefit: Improving Safety

Safety data sheets exist to prevent injuries. With the whole set searchable in one place you can:

- ⚠️ **Build a hazard register**: list every product with its signal word, hazard statements and pictograms
- 🥽 **Standardize PPE**: derive glove, eye, respiratory and skin protection requirements per product (Section 8)
- 🚑 **Create first-aid cards**: print one-page emergency summaries for site first-aiders (Section 4)
- 🔥 **Plan for fire and spills**: pull extinguishing media and spill procedures into your emergency plans (Sections 5 and 6)
- 📦 **Check storage compatibility**: use Sections 7 and 10 to avoid storing incompatible products together
- 🌬️ **Spot dust and silica risks**: cement-based products commonly list respirable crystalline silica, which has strict exposure limits. Search for it and prioritize controls
- 🎓 **Train workers**: turn real SDS content into toolbox-talk material
- 📋 **Support HazCom compliance**: in the US, OSHA's Hazard Communication Standard (29 CFR 1910.1200) requires SDSs to be readily accessible to workers. An organized, searchable library helps you meet that expectation (always confirm against the current official versions)

---

## ✅ Benefit: Improving Quality

Technical documents also drive installation quality and fewer callbacks:

- 📐 **Compare coverage and mixing ratios** so estimates and orders are accurate
- ⏱️ **Check pot life, open time and cure times** before scheduling trades
- 🌡️ **Verify temperature and substrate limits** to avoid failures from poor conditions
- 🏅 **Confirm standard references** (for example ANSI, ISO or ASTM listings) when writing project specifications
- 🧱 **Match the right product to the substrate and use** by comparing TDS performance values side by side
- 🔁 **Standardize procedures**: build internal checklists from installation guides so every crew works to the same method
- 🕵️ **Failure analysis**: when a job goes wrong, compare what was done against the document requirements to find the root cause

---

## 🔬 Benefit: Reverse Engineering & Product Analysis

"Reverse engineering" here means **analyzing published documents to understand how a product is built, classified and meant to perform**, not copying a proprietary formula.

What the documents can reveal:

- 🧪 **Ingredient families**: Section 3 lists hazardous ingredients, their CAS numbers and concentration **ranges** (for example cement, quartz sand, polymers)
- ⚖️ **Product type and chemistry**: hazard statements, pH and physical properties point to cementitious, epoxy, acrylic or other chemistries
- 🌿 **Environmental profile**: VOC content, ecological and disposal sections support green-building comparisons
- 🧭 **Competitive benchmarking**: compare ingredient classes, hazard classifications and stated performance across product lines
- 🧩 **System understanding**: see how primers, mortars, grouts and sealers are designed to work together
- 🔎 **Substitution and compatibility checks**: identify similar products when a specified item is unavailable (then verify with the manufacturer)
- 📉 **Trend analysis**: compare SDS revision dates to see how hazard classifications and ingredients change over time

> 🚧 **Important limits of this approach:**
>
> - SDSs only have to list **hazardous** components above reporting thresholds. Non-hazardous ingredients, additives and exact percentages are usually **not disclosed**, and trade secrets can be withheld.
> - You **cannot** recreate a manufacturer's formulation from an SDS, and you should not try to copy proprietary products. Respect intellectual property and the manufacturer's rights.
> - Treat any inference as a starting point for questions to the manufacturer, not as fact.

---

## 📊 More Use Cases

| Use Case                        | How This Archive Helps                                                                        |
| ------------------------------- | --------------------------------------------------------------------------------------------- |
| 🏢 **Contractors & installers** | Offline job-site reference for SDSs and installation steps                                    |
| 🧑‍💼 **Safety managers**          | Central chemical register and training material                                               |
| 📐 **Architects & specifiers**  | Quick access to standards references and system details                                       |
| 🎓 **Students & researchers**   | Real-world documents for studying construction chemistry and regulation                       |
| 🤖 **Data scientists / NLP**    | Corpus for text extraction, classification and entity recognition (CAS numbers, hazard codes) |
| 🗃️ **Document archivists**      | Preserve versions of documents that might be removed or updated online                        |
| 🛒 **Procurement teams**        | Compare products and keep compliance paperwork on file                                        |

---

## ⚠️ Limitations

- 🕰️ **Documents change.** Manufacturers update SDSs and data sheets. This archive is a **snapshot** and may be out of date.
- 🔄 **Skipping existing files** means an updated document with the **same file name** will not be re-downloaded. Delete the old file to refresh it.
- 📛 **Files are saved by URL file name.** Two different URLs ending in the same name will collide.
- 🧩 **Text extraction is imperfect**, especially for scanned PDFs or complex tables. Always verify important values in the original PDF.
- 🌍 **Regional differences.** Documents may be region-specific (for example US-focused regulations).

---

## ⚖️ Legal & Disclaimer

- 🏷️ All documents in this repository are the **property of their respective owner** (LATICRETE International, Inc. and/or its affiliates). This project is **not affiliated with, endorsed by, or sponsored by** them.
- 📚 This repository is provided for **archival, educational and research purposes**.
- 🦺 **Do not rely on this archive for safety-critical decisions.** Always obtain the **current** SDS and technical documents from the manufacturer or its official website before using, storing, transporting or specifying a product.
- 🤖 If you run the downloader, be respectful: keep the built-in delay between requests, and review the website's terms of use and `robots.txt`.
- 📝 The downloader code is free to use under the license you choose for this repository (for example MIT). Add a `LICENSE` file to make that explicit. The license applies to the **code only**, not to the manufacturer's documents.
- ⚕️ Nothing here is professional, legal or medical advice.

---

## 🤝 Contributing

Ideas and improvements are welcome! 🙌

1. 🍴 Fork the repository
2. 🌿 Create a branch: `git checkout -b feature/my-improvement`
3. 💾 Commit your changes: `git commit -m "Add my improvement"`
4. 🚀 Push and open a Pull Request

**Good first ideas:**

- 🧾 Add a script that builds a CSV index of every document
- 🏷️ Auto-classify documents (SDS, TDS, guide) from file names and content
- 📅 Record the download date and source URL for every file
- 🔔 Detect and flag changed documents on re-runs

---

## 👤 Author

**Prajwal Koirala**

- 🐙 GitHub: [@PrajwalKoirala638](https://github.com/PrajwalKoirala638)

---

⭐ If this project helps you, consider giving it a star! 🌟
