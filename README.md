# 🗂 Distributed File Management System (DFMS)

A secure and scalable distributed file sharing and storage system built with Go.  
DFMS supports file encryption, chunk-based storage, audit logging, and recovery from node failures.

---

## 📦 Project Overview

This project demonstrates how large files can be securely stored across distributed nodes using:

- ✅ **Hybrid Encryption (RSA + AES-GCM)**  
- ✅ **File Chunking & Distribution**  
- ✅ **Audit Logging**  
- ✅ **Failure Recovery**  
- ✅ **Public/Private Key Management**

---

## 📁 Folder Structure

```
FDS/
├── audit/               # Audit logging system
├── bootstrap/           # Node bootstrapping logic
├── chunks/              # Stored file chunks and metadata
├── crypto/              # Encryption and key management
├── keys/                # RSA key pairs
├── main.go              # Main server entry point
├── go.mod               # Go module definition
├── Makefile             # Build script
├── message.txt          # Sample message for testing
├── recovered_message.enc# Reconstructed file output
├── audit.log            # Log file with all events
```

---

## 🔐 Features

- **Hybrid Encryption (RSA + AES)** – Ensures data confidentiality.
- **Chunk-Based Storage** – Splits files into encrypted parts for efficient distribution.
- **Distributed Nodes** – Each chunk is stored separately for scalability.
- **Audit Logging** – Tracks all file operations in `audit.log`.
- **Fault Tolerance** – Automatically handles chunk loss or transfer failure.
- **Key Management** – Generates and handles secure RSA key pairs.

---

## 🚀 Getting Started

### ✅ Requirements

- Go 1.20 or newer
- OpenSSL (for key generation)

---

### 🛠️ Setup Instructions

1. **Clone the Repository**

```bash
git clone https://github.com/your-username/dfms.git
cd dfms/FDS
```

2. **Generate RSA Keys (if not present)**

```bash
mkdir -p keys/master
openssl genrsa -out keys/master/private.pem 2048
openssl rsa -in keys/master/private.pem -pubout -out keys/master/public.pem
```

3. **Build the Project**

```bash
make
```

4. **Run the System**

```bash
go run main.go
```

---

## 📤 Sending a File

1. Place your input file in the `FDS/` directory (e.g., `message.txt`).
2. Open `main.go` and specify the filename to be encrypted and shared.
3. The file will be:
   - Encrypted with a symmetric AES key.
   - AES key encrypted with RSA.
   - Chunked and distributed across the system.

---

## 📥 Receiving a File

- DFMS will automatically:
  - Reassemble file chunks.
  - Decrypt the final message.
- Output will be saved as: `recovered_message.enc`

---

## 🔒 Encryption Module Details

Located in `crypto/`:

| File               | Purpose                               |
|--------------------|----------------------------------------|
| `crypto.go`        | Core encryption/decryption logic       |
| `crypto_file.go`   | File-based encryption operations       |
| `crypto_secure.go` | AES key handling and secure transfer   |
| `keygen.go`        | RSA key generation logic               |

---

## 📚 Logging

All operations (upload/download/errors) are logged in:

```bash
audit.log
```

This helps in:
- Tracking system usage
- Debugging
- File access history

---

## 🤝 Contributing

We welcome contributions! Here's how you can help:
- 🐞 Fix bugs
- 🌟 Suggest features
- 🧪 Improve testing
- 📚 Enhance documentation

To contribute:
1. Fork the repo
2. Create a new branch (`git checkout -b feature-name`)
3. Make changes
4. Push and open a Pull Request

---

