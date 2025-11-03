package utils

import "strings"

// 根据后缀设置更准确的 ContentType
func GetContentType(fileSuffix string) string {
    
    suffix := strings.TrimPrefix(strings.ToLower(fileSuffix), ".")
    
    switch suffix {
    // 图片格式
    case "jpg", "jpeg":
        return "image/jpeg"
    case "png":
        return "image/png"
    case "gif":
        return "image/gif"
    case "bmp":
        return "image/bmp"
    case "webp":
        return "image/webp"
    case "svg", "svgz":
        return "image/svg+xml"
    case "ico":
        return "image/x-icon"
    case "tif", "tiff":
        return "image/tiff"
        
    // 文档格式
    case "pdf":
        return "application/pdf"
    case "doc":
        return "application/msword"
    case "docx":
        return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
    case "xls":
        return "application/vnd.ms-excel"
    case "xlsx":
        return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    case "ppt":
        return "application/vnd.ms-powerpoint"
    case "pptx":
        return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
        
    // 文本格式
    case "txt":
        return "text/plain"
    case "csv":
        return "text/csv"
    case "html", "htm":
        return "text/html"
    case "css":
        return "text/css"
    case "js":
        return "application/javascript"
    case "json":
        return "application/json"
    case "xml":
        return "application/xml"
        
    // 压缩格式
    case "zip":
        return "application/zip"
    case "rar":
        return "application/x-rar-compressed"
    case "7z":
        return "application/x-7z-compressed"
    case "tar":
        return "application/x-tar"
    case "gz":
        return "application/gzip"
        
    // 音视频格式
    case "mp3":
        return "audio/mpeg"
    case "wav":
        return "audio/wav"
    case "ogg":
        return "audio/ogg"
    case "mp4":
        return "video/mp4"
    case "avi":
        return "video/x-msvideo"
    case "mov":
        return "video/quicktime"
    case "webm":
        return "video/webm"
    case "flv":
        return "video/x-flv"
        
    // 编程相关
    case "go":
        return "text/plain; charset=utf-8"
    case "java":
        return "text/x-java-source"
    case "py":
        return "text/x-python"
    case "c", "h":
        return "text/x-c"
    case "cpp", "cc", "cxx", "hpp":
        return "text/x-c++"
    case "php":
        return "application/x-httpd-php"
    case "rb":
        return "application/x-ruby"
        
    // 其他常见格式
    case "rtf":
        return "application/rtf"
    case "epub":
        return "application/epub+zip"
    case "exe":
        return "application/x-msdownload"
    case "dmg":
        return "application/x-apple-diskimage"
    case "apk":
        return "application/vnd.android.package-archive"
        
    default:
        return "application/octet-stream"
    }
}