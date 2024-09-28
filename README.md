
**Compress-Master** is a powerful and efficient compression tool implemented in Go, leveraging the LZ77 and Huffman coding algorithms. Designed for both simplicity and performance, Compress-Master enables users to compress and decompress files seamlessly, offering customizable parameters and diagnostic outputs to optimize compression ratios and understand the underlying processes.

---

## Table of Contents

- [Compress-Master](#compress-master)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Installation](#installation)
    - [Prerequisites](#prerequisites)
    - [Building from Source](#building-from-source)
    - [Using Precompiled Binaries](#using-precompiled-binaries)
  - [Usage](#usage)
    - [Compression](#compression)
    - [Decompression](#decompression)
    - [Examples](#examples)
  - [Command-Line Flags](#command-line-flags)
  - [Dependencies](#dependencies)
  - [Profiling](#profiling)
  - [Testing](#testing)
  - [Contributing](#contributing)
  - [License](#license)
  - [Acknowledgments](#acknowledgments)
  - [Contact](#contact)

---

## Overview

Compress-Master utilizes the **LZ77** algorithm to identify and replace repeated sequences in the input data with references (pointers), followed by **Huffman coding** to encode the compressed data efficiently. This two-step approach ensures high compression ratios while maintaining speed and reliability.

Whether you're looking to save storage space, reduce transmission times, or simply explore compression algorithms, Compress-Master offers a robust solution tailored to your needs.

---

## Features

- **LZ77 Compression:** Efficiently identifies and replaces repeated sequences in data with pointers.
- **Huffman Coding:** Encodes the compressed data to minimize the overall size.
- **Customizable Parameters:**
  - **Minimum Match Length (`-min-match`):** Sets the smallest sequence length to consider for compression.
  - **Maximum Match Length (`-max-match`):** Sets the largest sequence length to consider.
  - **Search Buffer Size (`-search-size`):** Defines the size of the search window for identifying matches.
- **Diagnostic Outputs:**
  - **Huffman Tree Visualization (`-graphviz`):** Generates a Graphviz `.dot` file representing the Huffman tree.
  - **LZ77 Representation (`-lz`):** Outputs the LZ77 compressed sequence for analysis.
- **Performance Profiling (`-cpuprofile`):** Enables CPU profiling to analyze and optimize performance.
- **Verbose Logging (`-verbose`):** Provides detailed logs of the compression and decompression processes.
- **Cross-Platform Support:** Compatible with Windows, macOS, and Linux.

---

## Installation

### Prerequisites

Before you begin, ensure you have met the following requirements:

- **Go Programming Language:**  
  Ensure you have Go installed on your system. You can download it from the [official Go website](https://golang.org/dl/).

- **Git:**  
  Required for cloning the repository. Download it from the [official Git website](https://git-scm.com/downloads).

- **Graphviz (Optional):**  
  Needed for visualizing the Huffman tree if you utilize the `-graphviz` flag. Download it from the [Graphviz website](https://graphviz.org/download/).

### Building from Source

1. **Clone the Repository:**

   ```sh
   git clone https://github.com/OriLipper/compress-master.git
   cd compress-master
   ```

2. **Initialize Go Modules and Download Dependencies:**

   ```sh
   go mod tidy
   ```

3. **Build the Executable:**

   ```sh
   go build -o compress-master
   ```

   **Output:**
   This command generates an executable named `compress-master` (or `compress-master.exe` on Windows) in your project directory.

### Using Precompiled Binaries

If you prefer not to build from source, you can download precompiled binaries from the Releases page of the repository. Follow the instructions provided there for your specific operating system.

---

## Usage

Compress-Master operates in two primary modes: compression and decompression. Depending on your needs, you can customize the behavior using various command-line flags.

### Compression

Compress a file using LZ77 and Huffman coding.

**Basic Command Structure:**

```sh
./compress-master -compress=true [OPTIONS] <input_file>
```

**Example:**

```sh
./compress-master -compress=true -name=dickens.txt.compressed -verbose=true -graphviz=tree.dot -lz=lz_output.txt dickens.txt
```

**Parameters:**

- `-compress=true`: Sets the program to compression mode (default mode).
- `-name=dickens.txt.compressed`: Specifies the name of the compressed output file.
- `-verbose=true`: Enables verbose logging for detailed output.
- `-graphviz=tree.dot`: Outputs the Huffman tree visualization to `tree.dot`.
- `-lz=lz_output.txt`: Outputs the LZ77 representation to `lz_output.txt`.
- `dickens.txt`: The input file to compress.

**Outcome:**

- Compressed File: `dickens.txt.compressed`
- Huffman Tree Visualization: `tree.dot`
- LZ77 Representation: `lz_output.txt`

### Decompression

Decompress a previously compressed file.

**Basic Command Structure:**

```sh
./compress-master -compress=false [OPTIONS] <compressed_file>
```

**Example:**

```sh
./compress-master -compress=false -name=dickens_uncompressed.txt -verbose=true dickens.txt.compressed
```

**Parameters:**

- `-compress=false`: Sets the program to decompression mode.
- `-name=dickens_uncompressed.txt`: Specifies the name of the decompressed output file.
- `-verbose=true`: Enables verbose logging for detailed output.
- `dickens.txt.compressed`: The compressed file to decompress.

**Outcome:**

- Decompressed File: `dickens_uncompressed.txt`

---

## Command-Line Flags

Compress-Master offers a variety of command-line flags to customize its behavior. Below is a comprehensive list of available flags and their descriptions.

| Flag          | Type  | Default Value | Description                                                                                       |
| ------------- | ----- | ------------- | ------------------------------------------------------------------------------------------------- |
| `-compress`   | bool  | true          | Mode Selector: Set to true for compression and false for decompression. Default is compression mode. |
| `-name`       | string | `<input_file>.compressed` or `<input_file>.decompressed` | Specifies the name of the output file. If omitted, the program appends `.compressed` or `.decompressed` to the input filename based on the mode. |
| `-min-match`  | uint  | 4             | LZ77 Parameter: Sets the minimum match length for the LZ77 algorithm.                               |
| `-max-match`  | uint  | 255           | LZ77 Parameter: Sets the maximum match length for the LZ77 algorithm.                               |
| `-search-size`| uint  | 4096          | LZ77 Parameter: Defines the size of the search window for the LZ77 algorithm.                       |
| `-verbose`    | bool  | false         | Enables verbose logging to display detailed process information.                                   |
| `-graphviz`   | string | "" (empty)    | Outputs the Huffman tree visualization to the specified `.dot` file. Useful for generating graphical representations using Graphviz tools. |
| `-lz`         | string | "" (empty)    | Outputs the LZ77 representation of the compressed data to the specified file. Useful for analysis and debugging of the compression process. |
| `-cpuprofile` | string | "" (empty)    | Enables CPU profiling and writes the profile data to the specified file. Useful for performance analysis and optimization. |

---

## Dependencies

Compress-Master relies on the following external packages:

- `github.com/icza/bitio v1.1.0`: Provides bit-level I/O operations, essential for efficient compression and decompression.

These dependencies are managed automatically via Go Modules. Running `go mod tidy` ensures that all necessary packages are downloaded and available.

---

## Profiling

Understanding the performance characteristics of your compression and decompression processes can be invaluable for optimization. Compress-Master supports CPU profiling, allowing you to analyze and improve its performance.

