package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Name    string
	Size    int64
	ModTime string
	IsDir   bool
	Path    string
	SizeStr string
}

type PageData struct {
	CurrentPath string
	Files       []FileInfo
	ParentPath  string
}

var basePath string

const htmlTemplate = `
<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>파일 서버 - {{.CurrentPath}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: rgba(255, 255, 255, 0.95);
            border-radius: 20px;
            padding: 30px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
            backdrop-filter: blur(10px);
        }
        
        .header {
            text-align: center;
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 2px solid #e0e0e0;
        }
        
        .header h1 {
            color: #333;
            font-size: 2.5rem;
            margin-bottom: 10px;
            background: linear-gradient(45deg, #667eea, #764ba2);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        
        .path-nav {
            background: #f8f9fa;
            padding: 15px 20px;
            border-radius: 10px;
            margin-bottom: 25px;
            font-family: 'Courier New', monospace;
            border-left: 4px solid #667eea;
        }
        
        .path-nav strong {
            color: #495057;
        }
        
        .back-button {
            display: inline-block;
            background: linear-gradient(45deg, #667eea, #764ba2);
            color: white;
            padding: 10px 20px;
            text-decoration: none;
            border-radius: 25px;
            margin-bottom: 20px;
            transition: all 0.3s ease;
            font-weight: 500;
        }
        
        .back-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
        }
        
        .file-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        
        .file-item {
            background: white;
            border-radius: 15px;
            padding: 20px;
            border: 1px solid #e0e0e0;
            transition: all 0.3s ease;
            position: relative;
            overflow: hidden;
        }
        
        .file-item:hover {
            transform: translateY(-5px);
            box-shadow: 0 15px 30px rgba(0, 0, 0, 0.1);
            border-color: #667eea;
        }
        
        .file-icon {
            font-size: 2.5rem;
            margin-bottom: 10px;
            display: block;
        }
        
        .folder-icon { color: #ffc107; }
        .file-icon-default { color: #6c757d; }
        .image-icon { color: #28a745; }
        .document-icon { color: #007bff; }
        .archive-icon { color: #dc3545; }
        
        .file-name {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
            word-break: break-word;
            font-size: 1.1rem;
        }
        
        .file-details {
            color: #6c757d;
            font-size: 0.9rem;
            margin-bottom: 15px;
        }
        
        .file-actions {
            display: flex;
            gap: 10px;
        }
        
        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 20px;
            text-decoration: none;
            font-size: 0.9rem;
            font-weight: 500;
            transition: all 0.3s ease;
            cursor: pointer;
            display: inline-flex;
            align-items: center;
            gap: 5px;
        }
        
        .btn-primary {
            background: linear-gradient(45deg, #667eea, #764ba2);
            color: white;
        }
        
        .btn-outline {
            background: transparent;
            color: #667eea;
            border: 2px solid #667eea;
        }
        
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(0, 0, 0, 0.2);
        }
        
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #6c757d;
        }
        
        .empty-state .icon {
            font-size: 4rem;
            margin-bottom: 20px;
            opacity: 0.5;
        }
        
        @media (max-width: 768px) {
            .file-grid {
                grid-template-columns: 1fr;
            }
            
            .container {
                padding: 20px;
                margin: 10px;
            }
            
            .header h1 {
                font-size: 2rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📁 파일 서버</h1>
            <p>원하는 파일을 선택하여 다운로드하세요</p>
        </div>
<!--
        <div class="path-nav">
            <strong>현재 경로:</strong> {{.CurrentPath}}
        </div>
-->
        {{if .ParentPath}}
        <a href="{{.ParentPath}}" class="back-button">
            ⬅ 상위 폴더로
        </a>
        {{end}}
        
        {{if .Files}}
        <div class="file-grid">
            {{range .Files}}
            <div class="file-item">
                {{if .IsDir}}
                    <span class="file-icon folder-icon">📁</span>
                    <div class="file-name">{{.Name}}</div>
                    <div class="file-details">폴더 • {{.ModTime}}</div>
                    <div class="file-actions">
                        <a href="/browse{{.Path}}" class="btn btn-primary">열기</a>
                    </div>
                {{else}}
                    <span class="file-icon file-icon-default">
                       📄
                    </span>
                    <div class="file-name">{{.Name}}</div>
                    <div class="file-details">
                        {{.SizeStr}} • {{.ModTime}}
                    </div>
                    <div class="file-actions">
                        <a href="/download{{.Path}}" class="btn btn-primary" download>다운로드</a>
                        <a href="/view{{.Path}}" class="btn btn-outline" target="_blank">미리보기</a>
                    </div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{else}}
        <div class="empty-state">
            <div class="icon">📂</div>
            <h3>폴더가 비어있습니다</h3>
            <p>이 폴더에는 파일이나 하위 폴더가 없습니다.</p>
        </div>
        {{end}}
    </div>
</body>
</html>
`

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	if size >= GB {
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	} else if size >= MB {
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	} else if size >= KB {
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	}
	return fmt.Sprintf("%d B", size)
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	requestPath := strings.TrimPrefix(r.URL.Path, "/browse")
	if requestPath == "" {
		requestPath = "/"
	}

	fullPath := filepath.Join(basePath, requestPath)

	// 보안 검사: basePath 외부로 나가는 것을 방지
	if !strings.HasPrefix(fullPath, basePath) {
		http.Error(w, "접근 거부", http.StatusForbidden)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "파일을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	if !info.IsDir() {
		http.Error(w, "디렉토리가 아닙니다", http.StatusBadRequest)
		return
	}

	files, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, "디렉토리를 읽을 수 없습니다", http.StatusInternalServerError)
		return
	}

	var fileInfos []FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(requestPath, file.Name())
		if requestPath == "/" {
			filePath = "/" + file.Name()
		}

		fileInfo := FileInfo{
			Name:    file.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
			IsDir:   file.IsDir(),
			Path:    filePath,
			SizeStr: formatSize(info.Size()),
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	// 부모 경로 계산
	var parentPath string
	if requestPath != "/" {
		parentPath = "/browse" + filepath.Dir(requestPath)
		if parentPath == "/browse/" {
			parentPath = "/browse"
		}
	}

	data := PageData{
		CurrentPath: fullPath,
		Files:       fileInfos,
		ParentPath:  parentPath,
	}

	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "템플릿 오류", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	requestPath := strings.TrimPrefix(r.URL.Path, "/download")
	fullPath := filepath.Join(basePath, requestPath)

	// 보안 검사
	if !strings.HasPrefix(fullPath, basePath) {
		http.Error(w, "접근 거부", http.StatusForbidden)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "파일을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		http.Error(w, "디렉토리는 다운로드할 수 없습니다", http.StatusBadRequest)
		return
	}

	file, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, "파일을 열 수 없습니다", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filename := filepath.Base(fullPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	io.Copy(w, file)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	requestPath := strings.TrimPrefix(r.URL.Path, "/view")
	fullPath := filepath.Join(basePath, requestPath)

	// 보안 검사
	if !strings.HasPrefix(fullPath, basePath) {
		http.Error(w, "접근 거부", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, fullPath)
}

func myIp() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	// IP를 찾지 못했을 때 처리
	return "IP를 찾을 수 없습니다"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("사용법: go run main.go <폴더경로> [포트번호]")
		fmt.Println("예시: go run main.go /home/user/documents 8080")
		os.Exit(1)
	}

	var err error
	basePath, err = filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatal("경로 오류:", err)
	}

	// 경로가 존재하는지 확인
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		log.Fatal("지정된 경로가 존재하지 않습니다:", basePath)
	}

	port := "8080"
	if len(os.Args) >= 3 {
		port = os.Args[2]
	}

	myLocalIP := myIp()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/browse/", http.StatusMovedPermanently)
	})
	http.HandleFunc("/browse", browseHandler)
	http.HandleFunc("/browse/", browseHandler)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/view/", viewHandler)

	fmt.Printf("🚀 파일 서버가 시작되었습니다!\n")
	fmt.Printf("📁 서빙 경로: %s\n", basePath)
	fmt.Printf("🌐 접속 주소: http://localhost:%s\n", port)
	fmt.Printf("%s:%s\n", myLocalIP, port)
	fmt.Printf("⏹️  종료하려면 Ctrl+C를 누르세요\n\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
