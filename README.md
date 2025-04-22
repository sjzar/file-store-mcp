<div align="center">

# File Store MCP Server

![File Store MCP](https://socialify.git.ci/sjzar/file-store-mcp/image?font=Rokkitt&name=1&pattern=Diagonal+Stripes&theme=Auto)

_A file storage service supporting multiple cloud providers_

[![Go Report Card](https://goreportcard.com/badge/github.com/sjzar/file-store-mcp)](https://goreportcard.com/report/github.com/sjzar/file-store-mcp)
[![GoDoc](https://godoc.org/github.com/sjzar/file-store-mcp?status.svg)](https://godoc.org/github.com/sjzar/file-store-mcp)
[![GitHub release](https://img.shields.io/github/release/sjzar/file-store-mcp.svg)](https://github.com/sjzar/file-store-mcp/releases)
[![GitHub license](https://img.shields.io/github/license/sjzar/file-store-mcp.svg)](https://github.com/sjzar/file-store-mcp/blob/main/LICENSE)

</div>

## Features

- Multi-cloud storage provider support
- Unified API for file uploads
- Presigned URL generation for secure access
- Support for AWS S3, Alibaba Cloud OSS, Tencent Cloud COS, Qiniu Cloud, and GitHub
- Easy configuration via environment variables
- Customizable URL expiration times
- Support for custom domains and CDNs
- Dual operation modes: stdio for direct integration and SSE for server mode

## Quick Start

### Installation

```bash
go install github.com/sjzar/file-store-mcp@latest
```

### Basic Usage

1. Set the required environment variables for your chosen storage provider
2. Run in stdio mode (default) for direct integration with other applications:
   ```bash
   file-store-mcp
   ```
3. Or run in SSE server mode for HTTP-based integration:
   ```bash
   file-store-mcp --sse-port 8080
   ```


## MCP Tools

File Store MCP provides three tools for uploading files to cloud storage:

### 1. Upload Files Tool (`upload_files`)

Uploads local files to cloud storage and returns HTTP URLs.

**When to use**: When users mention local file paths or need online access to their files. Ideal for analyzing PDF content, referencing local images for drawing tasks, or processing any local files.

**Parameters**:
- `paths`: Array of absolute local file paths to upload (required)

**Example**:
```json
{
  "tool": "upload_files",
  "params": {
    "paths": ["/path/to/file1.pdf", "C:/Users/user/Documents/file2.jpg"]
  }
}
```

### 2. Upload Clipboard Files Tool (`upload_clipboard_files`)

Uploads files from the clipboard to cloud storage and returns HTTP URLs.

**When to use**: Only when users explicitly request to upload files from their clipboard. Useful when users want to share or process clipboard content without saving it locally first.

**Parameters**: None required

**Example**:
```json
{
  "tool": "upload_clipboard_files"
}
```

### 3. Upload URL Files Tool (`upload_url_files`)

Downloads files from provided URLs and uploads them to cloud storage, returning new HTTP URLs.

**When to use**: When users provide web links to files they want to process or analyze. Ideal for situations where users reference external files that need to be incorporated into the current workflow.

**Parameters**:
- `urls`: Array of URLs pointing to files to download and upload (required)

**Example**:
```json
{
  "tool": "upload_url_files",
  "params": {
    "urls": ["https://example.com/file1.pdf", "https://another-site.org/image.jpg"]
  }
}
```

## Storage Providers

File Store MCP supports the following storage providers:

- AWS S3 (and compatible services like Cloudflare R2)
- Alibaba Cloud OSS
- Tencent Cloud COS
- Qiniu Cloud Storage
- GitHub Repository

## Configuration

### Common Configuration

| Environment Variable | Description | Default |
|----------------------|-------------|---------|
| `FSM_STORAGE_TYPE` | Storage provider type (s3, oss, cos, qiniu, github) | `empty` |

### AWS S3 Configuration

Set `FSM_STORAGE_TYPE=s3` to use AWS S3 or compatible services.

| Environment Variable | Description | Required | Default |
|----------------------|-------------|----------|---------|
| `FSM_S3_BUCKET` | S3 bucket name | Yes | - |
| `FSM_S3_REGION` | AWS region | Yes | - |
| `FSM_S3_ENDPOINT` | Custom endpoint for S3-compatible services | No | AWS S3 endpoint |
| `FSM_S3_ACCESS_KEY` | AWS access key ID | Yes | - |
| `FSM_S3_SECRET_KEY` | AWS secret access key | Yes | - |
| `FSM_S3_SESSION` | AWS session token | No | - |
| `FSM_S3_URL_EXPIRATION` | Presigned URL expiration time in seconds | No | 604800 (7 days) |

**Notes for S3-compatible services:**
- For Cloudflare R2: Set `FSM_S3_ENDPOINT` to your R2 endpoint URL
- For other S3-compatible services: Configure the appropriate endpoint URL

### Alibaba Cloud OSS Configuration

Set `FSM_STORAGE_TYPE=oss` to use Alibaba Cloud OSS.

| Environment Variable | Description | Required | Default |
|----------------------|-------------|----------|---------|
| `FSM_OSS_ENDPOINT` | OSS endpoint | Yes | - |
| `FSM_OSS_ACCESS_KEY` | OSS access key ID | Yes | - |
| `FSM_OSS_SECRET_KEY` | OSS access key secret | Yes | - |
| `FSM_OSS_BUCKET` | OSS bucket name | Yes | - |
| `FSM_OSS_DOMAIN` | Custom domain for OSS bucket | No | - |
| `FSM_OSS_URL_EXPIRATION` | Signed URL expiration time in seconds | No | 604800 (7 days) |

### Tencent Cloud COS Configuration

Set `FSM_STORAGE_TYPE=cos` to use Tencent Cloud COS.

| Environment Variable | Description | Required | Default |
|----------------------|-------------|----------|---------|
| `FSM_COS_BUCKET` | COS bucket name | Yes | - |
| `FSM_COS_REGION` | COS region | Yes | - |
| `FSM_COS_APP_ID` | Tencent Cloud App ID | Yes | - |
| `FSM_COS_ACCESS_KEY` | Secret ID | Yes | - |
| `FSM_COS_SECRET_KEY` | Secret Key | Yes | - |
| `FSM_COS_DOMAIN` | Custom domain for COS bucket | No | - |
| `FSM_COS_USE_HTTPS` | Whether to use HTTPS | No | `true` |
| `FSM_COS_USE_ACCELERATE` | Whether to use global acceleration | No | `false` |
| `FSM_COS_URL_EXPIRATION` | Presigned URL expiration time in seconds | No | 604800 (7 days) |

### Qiniu Cloud Storage Configuration

Set `FSM_STORAGE_TYPE=qiniu` to use Qiniu Cloud Storage.

| Environment Variable | Description | Required | Default |
|----------------------|-------------|----------|---------|
| `FSM_QINIU_ACCESS_KEY` | Qiniu access key | Yes | - |
| `FSM_QINIU_SECRET_KEY` | Qiniu secret key | Yes | - |
| `FSM_QINIU_BUCKET` | Qiniu bucket name | Yes | - |
| `FSM_QINIU_DOMAIN` | Custom domain for Qiniu bucket (required) | Yes | - |
| `FSM_QINIU_REGION` | Storage region | No | `z0` (East China) |
| `FSM_QINIU_URL_EXPIRATION` | Signed URL expiration time in seconds | No | 604800 (7 days) |

**Available Qiniu regions:**
- `z0`: East China
- `z1`: North China
- `z2`: South China
- `na0`: North America
- `as0`: Southeast Asia

### GitHub Repository Configuration

Set `FSM_STORAGE_TYPE=github` to use GitHub as a storage provider.

| Environment Variable | Description | Required | Default |
|----------------------|-------------|----------|---------|
| `FSM_GITHUB_TOKEN` | GitHub personal access token | Yes | - |
| `FSM_GITHUB_OWNER` | Repository owner | Yes | - |
| `FSM_GITHUB_REPO` | Repository name | Yes | - |
| `FSM_GITHUB_BRANCH` | Branch name | No | `main` |
| `FSM_GITHUB_PATH` | File storage path within the repository | No | - |
| `FSM_GITHUB_DOMAIN` | Custom domain for GitHub content | No | - |

**GitHub token permissions:**
- The token must have `repo` scope for private repositories
- For public repositories, `public_repo` scope is sufficient

## Advanced Usage

### Using Custom Domains

For all storage providers, you can configure custom domains to serve your files. This is particularly useful when you have CDN services in front of your storage.

Example for Alibaba Cloud OSS with custom domain:
```
FSM_STORAGE_TYPE=oss
FSM_OSS_BUCKET=my-bucket
FSM_OSS_ENDPOINT=oss-cn-hangzhou.aliyuncs.com
FSM_OSS_ACCESS_KEY=your-access-key
FSM_OSS_SECRET_KEY=your-secret-key
FSM_OSS_DOMAIN=cdn.example.com
```

### Debug Mode

Enable debug mode for more verbose logging:

```bash
file-store-mcp --debug
```

## Development

### Building from Source

```bash
git clone https://github.com/sjzar/file-store-mcp.git
cd file-store-mcp
go build
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
