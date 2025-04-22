package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	Name = "file-store-mcp"
)

var UploadFilesTool = mcp.NewTool(
	"upload_files",
	mcp.WithDescription("Uploads local files to cloud storage and returns HTTP URLs. Use this tool when users mention local file paths or need online access to their files. Ideal for when users want to: analyze PDF content, reference local images for drawing tasks, or process any local files. If input contains absolute paths (like 'C:/Users/file.pdf', '/home/user/image.jpg'), use this tool to obtain web-accessible links."),
	mcp.WithArray("paths", mcp.Description("array of absolute local file paths to upload"), mcp.Required()),
)

var UploadClipboardFilesTool = mcp.NewTool(
	"upload_clipboard_files",
	mcp.WithDescription("Uploads files from the clipboard to cloud storage and returns HTTP URLs. Only use this tool when users explicitly request to upload files from their clipboard. Useful when users want to share or process clipboard content without saving it locally first. This tool helps users easily convert clipboard files into web-accessible resources."),
)

var UploadUrlFilesTool = mcp.NewTool(
	"upload_url_files",
	mcp.WithDescription("Downloads files from provided URLs and uploads them to cloud storage, returning new HTTP URLs. Use this tool when users provide web links to files they want to process or analyze. Ideal for situations where users reference external files that need to be incorporated into the current workflow. This tool simplifies working with content from various online sources."),
	mcp.WithArray("urls", mcp.Description("array of URLs pointing to files to download and upload"), mcp.Required()),
)
