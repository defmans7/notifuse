package keygenfunction

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http"

	"aidanwoods.dev/go-paseto"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("KeygenFunction", KeygenHandler)
}

type KeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

// KeygenHandler serves both the HTML page and handles key generation
func KeygenHandler(w http.ResponseWriter, r *http.Request) {
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

	data := struct {
		LogoBase64 string
	}{
		LogoBase64: logoBase64,
	}

	tmpl := template.Must(template.New("keygen").Parse(htmlTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// generateKeys generates a new PASETO v4 key pair and returns it as JSON
func generateKeys(w http.ResponseWriter, r *http.Request) {
	// Generate a new key pair
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Convert keys to base64 for storage
	privateKeyBase64 := base64.StdEncoding.EncodeToString(secretKey.ExportBytes())
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey.ExportBytes())

	// Create response
	keyPair := KeyPair{
		PrivateKey: privateKeyBase64,
		PublicKey:  publicKeyBase64,
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

const logoBase64 = "iVBORw0KGgoAAAANSUhEUgAAAMgAAABMCAYAAAAoefhQAAAAAXNSR0IArs4c6QAAAERlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAA6ABAAMAAAABAAEAAKACAAQAAAABAAAAyKADAAQAAAABAAAATAAAAADfPSvHAAAU90lEQVR4Ae1dC3xUxbmfObt5IShP5WEJDxGQ4qOoFbk/e29vEqBkQ3ikCv60eZHC9W2tpYVro16Ver3aloc0kJCKYG0uQrLhkYSrvS2U2wqUqqC8wTYgkgICmk32nDP3P5uc5bw2u0k2u5tkzu+3e2a+ef9nvplvvvOdOZSISyDQhRG4NHL+tR6vnMkocRHGRjNCBlNKFELoKfg/lAitoPFJFf2OLL1oBwO1IwqaQKCzI3B29NO9mOfcjxgjTxDCerTcHnpOovSF/j2HLKP7Cxv1cQWD6NEQ7i6BwNkR825kiloB5hjdmgZRSv+Y5HTO6HV05edaOsEgGhLi3iUQqEvOH6syspMR1qdNDaL0eA+n8y6NSaQ2ZSISCQRiEIELQxf0AXO428wcvE2MDa+X5Y1sXGE89woG4SiIq0sg0Ei9z4M5Rra3MYyxu+su1z7M8xEiVnvRFOljAoEzNxSMpF71YwzuuPBUiJ5zxicNd4YnM5GLQCC6CEiNyv0qIWFiDt4W1pc11mcEZJDs7MLE0198OlpR2UBJYT55LLoQiNIFAkBAIp7qitIaMxaM0unYP5jJ7fKrhFkZJDVz3q1ElhfV1p2YiuKu4iWAM8UlEIgNBFRSi4pcvE2MDa+X5Y1sXGE89woG4SiIq0sg0Ei9z4M5Rra3MYyxu+su1z7M8xEiVnvRFOljAoEzNxSMpF71YwzuuPBUiJ5zxicNd4YnM5GLQCC6CEiNyv0qIWFiDt4W1pc11mcEZJDs7MLE0198OlpR2UBJYT55LLoQiNIFAkBAIp7qitIaMxaM0unYP5jJ7fKrhFkZJDVz3q1ElhfV1p2YiuKu4iWAM8UlEIgNBFRSi4pcvE2MDa+X5Y1sXGE89woG4SiIq0sg0Ei9z4M5Rra3MYyxu+su1z7M8xEiVnvRFOljAoEzNxSMpF71YwzuuPBUiJ5zxicNd4YnM5GLQCC6CEiNyv0qIWFiDt4W1pc11mcEZJDs7MLE0198OlpR2UBJYT55LLoQiNIFAkBAIp7qitIaMxaM0unYP5jJ7fKrhFkZJDVz3q1ElhfV1p2YiuKu4iWAM8UlEIgNBFRSi4pc"

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
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <a href="https://www.notifuse.com" target="_blank" class="logo-link">
                <img src="data:image/png;base64,{{.LogoBase64}}" alt="Notifuse" class="logo">
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
            <div class="key-section private">
                <h3>Private Key</h3>
                <div class="key-display" id="private-key">
                    <button class="copy-btn" onclick="copyToClipboard('private-key-text', this)">Copy</button>
                    <span id="private-key-text"></span>
                </div>
            </div>
            
            <div class="key-section public">
                <h3>Public Key</h3>
                <div class="key-display" id="public-key">
                    <button class="copy-btn" onclick="copyToClipboard('public-key-text', this)">Copy</button>
                    <span id="public-key-text"></span>
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
                
                // Display the keys
                document.getElementById('private-key-text').textContent = data.privateKey;
                document.getElementById('public-key-text').textContent = data.publicKey;
                
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
        
        // Generate keys on page load for demo
        window.addEventListener('load', () => {
            // Uncomment the line below if you want to auto-generate keys on page load
            // generateKeys();
        });
    </script>
</body>
</html>
`
