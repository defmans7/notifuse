package keygenfunction

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("KeygenFunction", KeygenHandler)
}

type KeyPair struct {
	PrivateKey  KeyData   `json:"privateKey"`
	PublicKey   KeyData   `json:"publicKey"`
	GeneratedAt time.Time `json:"generatedAt"`
	KeyType     string    `json:"keyType"`
	Algorithm   string    `json:"algorithm"`
}

type KeyData struct {
	Base64     string `json:"base64"`
	Hex        string `json:"hex"`
	ByteLength int    `json:"byteLength"`
}

// KeygenHandler serves both the HTML page and handles key generation
func KeygenHandler(w http.ResponseWriter, r *http.Request) {
	// Handle logo.png file serving
	if r.URL.Path == "/logo.png" && r.Method == http.MethodGet {
		serveLogo(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Serve the HTML page
		serveHTML(w, r)
	case http.MethodPost:
		// Handle key generation
		generateKeys(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// serveHTML serves the HTML page for key generation
func serveHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmpl := template.Must(template.New("keygen").Parse(htmlTemplate))
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// serveLogo serves the logo.png file as a static file
func serveLogo(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers for PNG
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Try different possible paths for the logo file
	possiblePaths := []string{
		"logo.png",
		"./logo.png",
		"serverless_function_source_code/logo.png",
		"/workspace/logo.png",
		"/tmp/logo.png",
	}

	var logoData []byte
	var err error

	for _, path := range possiblePaths {
		logoData, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		// Debug: list current directory contents
		files, _ := os.ReadDir(".")
		var fileList []string
		for _, file := range files {
			fileList = append(fileList, file.Name())
		}

		// Also check serverless_function_source_code directory
		sourceFiles, _ := os.ReadDir("serverless_function_source_code")
		var sourceFileList []string
		for _, file := range sourceFiles {
			sourceFileList = append(sourceFileList, file.Name())
		}

		http.Error(w, fmt.Sprintf("Logo not found. Current directory files: %v, Source directory files: %v", fileList, sourceFileList), http.StatusNotFound)
		return
	}

	w.Write(logoData)
}

// generateKeys generates a new PASETO v4 key pair and returns it as JSON
func generateKeys(w http.ResponseWriter, r *http.Request) {
	// Generate a new key pair
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Export raw bytes
	privateKeyBytes := secretKey.ExportBytes()
	publicKeyBytes := publicKey.ExportBytes()

	// Create response with multiple encoding formats
	keyPair := KeyPair{
		PrivateKey: KeyData{
			Base64:     base64.StdEncoding.EncodeToString(privateKeyBytes),
			Hex:        hex.EncodeToString(privateKeyBytes),
			ByteLength: len(privateKeyBytes),
		},
		PublicKey: KeyData{
			Base64:     base64.StdEncoding.EncodeToString(publicKeyBytes),
			Hex:        hex.EncodeToString(publicKeyBytes),
			ByteLength: len(publicKeyBytes),
		},
		GeneratedAt: time.Now().UTC(),
		KeyType:     "asymmetric",
		Algorithm:   "PASETO v4",
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Encode and send response
	if err := json.NewEncoder(w).Encode(keyPair); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PASETO v4 Key Generator</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
            padding: 40px;
            max-width: 800px;
            width: 100%;
        }
        
        .header {
            text-align: center;
            margin-bottom: 40px;
        }
        
        .logo-link {
            display: inline-block;
            margin-bottom: 20px;
            transition: opacity 0.2s ease;
        }
        
        .logo-link:hover {
            opacity: 0.8;
        }
        
        .logo {
            height: 60px;
            width: auto;
        }
        
        .header h1 {
            color: #333;
            font-size: 2.5rem;
            margin-bottom: 10px;
            font-weight: 600;
        }
        
        .header p {
            color: #666;
            font-size: 1.1rem;
        }
        
        .generate-section {
            text-align: center;
            margin-bottom: 40px;
        }
        
        .generate-btn {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 15px 30px;
            font-size: 1.1rem;
            border-radius: 8px;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
            font-weight: 600;
        }
        
        .generate-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
        }
        
        .generate-btn:active {
            transform: translateY(0);
        }
        
        .generate-btn:disabled {
            opacity: 0.6;
            cursor: not-allowed;
            transform: none;
        }
        
        .results {
            display: none;
            animation: fadeIn 0.5s ease-in-out;
        }
        
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }
        
        .key-section {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
            border-left: 4px solid #667eea;
        }
        
        .key-section h3 {
            color: #333;
            margin-bottom: 15px;
            font-size: 1.2rem;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .key-section.private h3::before {
            content: "🔐";
        }
        
        .key-section.public h3::before {
            content: "🔑";
        }
        
        .key-display {
            background: white;
            border: 1px solid #e9ecef;
            border-radius: 6px;
            padding: 15px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.9rem;
            word-break: break-all;
            line-height: 1.5;
            color: #495057;
            position: relative;
        }
        
        .copy-btn {
            position: absolute;
            top: 10px;
            right: 10px;
            background: #667eea;
            color: white;
            border: none;
            padding: 5px 10px;
            border-radius: 4px;
            font-size: 0.8rem;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        
        .copy-btn:hover {
            background: #5a6fd8;
        }
        
        .copy-btn.copied {
            background: #28a745;
        }
        
        .warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 6px;
            padding: 15px;
            margin-top: 20px;
            color: #856404;
        }
        
        .warning strong {
            color: #dc3545;
        }
        
        .loading {
            display: none;
            text-align: center;
            color: #667eea;
            font-size: 1.1rem;
        }
        
        .spinner {
            display: inline-block;
            width: 20px;
            height: 20px;
            border: 3px solid #f3f3f3;
            border-top: 3px solid #667eea;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-right: 10px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .error {
            display: none;
            background: #f8d7da;
            border: 1px solid #f5c6cb;
            border-radius: 6px;
            padding: 15px;
            margin-top: 20px;
            color: #721c24;
        }
        
        .key-metadata {
            background: #e3f2fd;
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 20px;
            border-left: 4px solid #2196f3;
        }
        
        .key-metadata p {
            margin: 5px 0;
            color: #1565c0;
        }
        
        .format-tabs {
            display: flex;
            margin-bottom: 10px;
            border-bottom: 1px solid #dee2e6;
        }
        
        .format-tab {
            background: transparent;
            border: none;
            padding: 8px 16px;
            cursor: pointer;
            border-bottom: 2px solid transparent;
            color: #666;
            font-weight: 500;
            transition: all 0.2s;
        }
        
        .format-tab:hover {
            color: #667eea;
        }
        
        .format-tab.active {
            color: #667eea;
            border-bottom-color: #667eea;
        }
        
        .key-info {
            margin-top: 8px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <a href="https://www.notifuse.com" target="_blank" class="logo-link">
                <img src="./logo.png" alt="Notifuse" class="logo">
            </a>
            <h1>PASETO v4 Key Generator</h1>
            <p>Generate secure asymmetric key pairs for PASETO v4 tokens</p>
        </div>
        
        <div class="generate-section">
            <button class="generate-btn" onclick="generateKeys()">
                Generate New Key Pair
            </button>
        </div>
        
        <div class="loading" id="loading">
            <div class="spinner"></div>
            Generating keys...
        </div>
        
        <div class="error" id="error">
            <strong>Error:</strong> <span id="error-message"></span>
        </div>
        
        <div class="results" id="results">
            <div class="key-metadata" id="key-metadata">
                <p><strong>Generated:</strong> <span id="generated-at"></span></p>
                <p><strong>Algorithm:</strong> <span id="algorithm"></span></p>
                <p><strong>Type:</strong> <span id="key-type"></span></p>
            </div>
            
            <div class="key-section private">
                <h3>Private Key</h3>
                <div class="format-tabs">
                    <button class="format-tab active" onclick="switchFormat('private', 'base64')">Base64</button>
                    <button class="format-tab" onclick="switchFormat('private', 'hex')">Hex</button>
                </div>
                <div class="key-display" id="private-key-base64">
                    <button class="copy-btn" onclick="copyToClipboard('private-key-base64-text', this)">Copy</button>
                    <span id="private-key-base64-text"></span>
                </div>
                <div class="key-display" id="private-key-hex" style="display: none;">
                    <button class="copy-btn" onclick="copyToClipboard('private-key-hex-text', this)">Copy</button>
                    <span id="private-key-hex-text"></span>
                </div>
                <div class="key-info">
                    <small>Length: <span id="private-key-length"></span> bytes</small>
                </div>
            </div>
            
            <div class="key-section public">
                <h3>Public Key</h3>
                <div class="format-tabs">
                    <button class="format-tab active" onclick="switchFormat('public', 'base64')">Base64</button>
                    <button class="format-tab" onclick="switchFormat('public', 'hex')">Hex</button>
                </div>
                <div class="key-display" id="public-key-base64">
                    <button class="copy-btn" onclick="copyToClipboard('public-key-base64-text', this)">Copy</button>
                    <span id="public-key-base64-text"></span>
                </div>
                <div class="key-display" id="public-key-hex" style="display: none;">
                    <button class="copy-btn" onclick="copyToClipboard('public-key-hex-text', this)">Copy</button>
                    <span id="public-key-hex-text"></span>
                </div>
                <div class="key-info">
                    <small>Length: <span id="public-key-length"></span> bytes</small>
                </div>
            </div>
            
            <div class="warning">
                <strong>⚠️ Security Warning:</strong> Keep your private key secret! Only share the public key. 
                Store the private key securely and never expose it in client-side code or public repositories.
            </div>
        </div>
    </div>

    <script>
        async function generateKeys() {
            const generateBtn = document.querySelector('.generate-btn');
            const loading = document.getElementById('loading');
            const results = document.getElementById('results');
            const error = document.getElementById('error');
            
            // Reset UI
            generateBtn.disabled = true;
            loading.style.display = 'block';
            results.style.display = 'none';
            error.style.display = 'none';
            
            try {
                const response = await fetch(window.location.href, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                });
                
                if (!response.ok) {
                    throw new Error(` + "`" + `HTTP error! status: ${response.status}` + "`" + `);
                }
                
                const data = await response.json();
                
                // Display metadata
                document.getElementById('generated-at').textContent = new Date(data.generatedAt).toLocaleString();
                document.getElementById('algorithm').textContent = data.algorithm;
                document.getElementById('key-type').textContent = data.keyType;
                
                // Display private key data
                document.getElementById('private-key-base64-text').textContent = data.privateKey.base64;
                document.getElementById('private-key-hex-text').textContent = data.privateKey.hex;
                document.getElementById('private-key-length').textContent = data.privateKey.byteLength;
                
                // Display public key data
                document.getElementById('public-key-base64-text').textContent = data.publicKey.base64;
                document.getElementById('public-key-hex-text').textContent = data.publicKey.hex;
                document.getElementById('public-key-length').textContent = data.publicKey.byteLength;
                
                // Show results
                loading.style.display = 'none';
                results.style.display = 'block';
                
            } catch (err) {
                console.error('Error generating keys:', err);
                loading.style.display = 'none';
                error.style.display = 'block';
                document.getElementById('error-message').textContent = err.message;
            } finally {
                generateBtn.disabled = false;
            }
        }
        
        async function copyToClipboard(elementId, button) {
            const text = document.getElementById(elementId).textContent;
            
            try {
                await navigator.clipboard.writeText(text);
                
                // Show feedback
                const originalText = button.textContent;
                button.textContent = 'Copied!';
                button.classList.add('copied');
                
                setTimeout(() => {
                    button.textContent = originalText;
                    button.classList.remove('copied');
                }, 2000);
                
            } catch (err) {
                console.error('Failed to copy text: ', err);
                
                // Fallback for older browsers
                const textArea = document.createElement('textarea');
                textArea.value = text;
                document.body.appendChild(textArea);
                textArea.focus();
                textArea.select();
                
                try {
                    document.execCommand('copy');
                    button.textContent = 'Copied!';
                    button.classList.add('copied');
                    
                    setTimeout(() => {
                        button.textContent = 'Copy';
                        button.classList.remove('copied');
                    }, 2000);
                } catch (fallbackErr) {
                    console.error('Fallback copy failed: ', fallbackErr);
                }
                
                document.body.removeChild(textArea);
            }
        }
        
        function switchFormat(keyType, format) {
            // Update tab states
            const tabs = document.querySelectorAll(` + "`" + `.key-section.${keyType} .format-tab` + "`" + `);
            tabs.forEach(tab => tab.classList.remove('active'));
            event.target.classList.add('active');
            
            // Show/hide content
            const base64Display = document.getElementById(` + "`" + `${keyType}-key-base64` + "`" + `);
            const hexDisplay = document.getElementById(` + "`" + `${keyType}-key-hex` + "`" + `);
            
            if (format === 'base64') {
                base64Display.style.display = 'block';
                hexDisplay.style.display = 'none';
            } else {
                base64Display.style.display = 'none';
                hexDisplay.style.display = 'block';
            }
        }
        
        // Generate keys on page load for demo
        window.addEventListener('load', () => {
            // Uncomment the line below if you want to auto-generate keys on page load
            // generateKeys();
        });
    </script>
</body>
</html>
`
