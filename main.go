package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	TargetPrintDPI   = 300
	MinimumAcceptDPI = 150
)

type PreflightDefect struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type ProofRecord struct {
	ID             string            `json:"id"`
	FileName       string            `json:"file_name"`
	TargetWidthIn  float64           `json:"target_width_in"`
	TargetHeightIn float64           `json:"target_height_in"`
	PixelWidth     int               `json:"pixel_width"`
	PixelHeight    int               `json:"pixel_height"`
	EffectiveDPI   int               `json:"effective_dpi"`
	Status         string            `json:"status"`
	CustomerNotes  string            `json:"customer_notes"`
	Defects        []PreflightDefect `json:"defects"`
	ProofImagePNG  []byte            `json:"-"`
	CreatedAt      time.Time         `json:"created_at"`
	DecidedAt      *time.Time        `json:"decided_at,omitempty"`
}

type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]*ProofRecord
}

var store = &MemoryStore{
	records: make(map[string]*ProofRecord),
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api/upload", handleUpload)
	http.HandleFunc("/api/proof/", handleViewProof)
	http.HandleFunc("/api/decision", handleDecision)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(portalHTML))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 25*1024*1024)
	if err := r.ParseMultipartForm(25 * 1024 * 1024); err != nil {
		http.Error(w, "File exceeds maximum upload size (25 MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("artwork")
	if err != nil {
		http.Error(w, "Invalid artwork file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	widthIn, _ := strconv.ParseFloat(r.FormValue("width_inches"), 64)
	heightIn, _ := strconv.ParseFloat(r.FormValue("height_inches"), 64)
	if widthIn <= 0 {
		widthIn = 4.0
	}
	if heightIn <= 0 {
		heightIn = 6.0
	}

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read upload data", http.StatusBadRequest)
		return
	}

	srcImg, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "The uploaded file is corrupt or unsupported (only PNG, JPEG, GIF supported).",
		})
		return
	}

	bounds := srcImg.Bounds()
	pixelW := bounds.Dx()
	pixelH := bounds.Dy()

	dpiX := float64(pixelW) / widthIn
	dpiY := float64(pixelH) / heightIn
	minDPI := int(dpiX)
	if int(dpiY) < minDPI {
		minDPI = int(dpiY)
	}

	var defects []PreflightDefect
	if minDPI < MinimumAcceptDPI {
		defects = append(defects, PreflightDefect{
			Severity: "CRITICAL",
			Message:  fmt.Sprintf("Resolution is %d DPI. This is below the minimum %d DPI and will print blurry.", minDPI, MinimumAcceptDPI),
		})
	} else if minDPI < TargetPrintDPI {
		defects = append(defects, PreflightDefect{
			Severity: "WARNING",
			Message:  fmt.Sprintf("Resolution is %d DPI. Recommended print standard is %d DPI.", minDPI, TargetPrintDPI),
		})
	}

	proofID := fmt.Sprintf("PRF-%d", time.Now().UnixNano()%1000000)
	proofBytes := buildProofImage(srcImg, minDPI, proofID)

	record := &ProofRecord{
		ID:             proofID,
		FileName:       header.Filename,
		TargetWidthIn:  widthIn,
		TargetHeightIn: heightIn,
		PixelWidth:     pixelW,
		PixelHeight:    pixelH,
		EffectiveDPI:   minDPI,
		Status:         "AWAITING_APPROVAL",
		Defects:        defects,
		ProofImagePNG:  proofBytes,
		CreatedAt:      time.Now(),
	}

	store.mu.Lock()
	store.records[proofID] = record
	store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func handleViewProof(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/proof/")
	store.mu.RLock()
	record, exists := store.records[id]
	store.mu.RUnlock()

	if !exists || len(record.ProofImagePNG) == 0 {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(record.ProofImagePNG)
}

func handleDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID       string `json:"id"`
		Decision string `json:"decision"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	record, exists := store.records[req.ID]
	if !exists {
		http.Error(w, "Proof ID not found", http.StatusNotFound)
		return
	}

	now := time.Now()
	record.Status = req.Decision
	record.CustomerNotes = req.Notes
	record.DecidedAt = &now

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func buildProofImage(src image.Image, dpi int, proofID string) []byte {
	bounds := src.Bounds()
	margin := 40
	w := bounds.Dx() + (margin * 2)
	h := bounds.Dy() + (margin * 2) + 60

	proof := image.NewRGBA(image.Rect(0, 0, w, h))

	bgColor := color.RGBA{245, 246, 248, 255}
	draw.Draw(proof, proof.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	artworkRect := image.Rect(margin, margin+40, margin+bounds.Dx(), margin+40+bounds.Dy())
	draw.Draw(proof, artworkRect, src, bounds.Min, draw.Over)

	// Draw Red Trim Cut Line
	red := color.RGBA{230, 57, 70, 255}
	drawHollowRect(proof, artworkRect.Min.X+8, artworkRect.Min.Y+8, artworkRect.Max.X-8, artworkRect.Max.Y-8, 2, red)

	// Draw Watermark Overlay Banner
	bannerY := (artworkRect.Min.Y + artworkRect.Max.Y) / 2
	bannerRect := image.Rect(margin, bannerY-24, margin+bounds.Dx(), bannerY+24)
	watermarkRed := color.RGBA{220, 50, 50, 180}
	draw.Draw(proof, bannerRect, &image.Uniform{watermarkRed}, image.Point{}, draw.Over)

	var buf bytes.Buffer
	png.Encode(&buf, proof)
	return buf.Bytes()
}

func drawHollowRect(dst *image.RGBA, x1, y1, x2, y2, thickness int, col color.Color) {
	for t := 0; t < thickness; t++ {
		for x := x1; x <= x2; x++ {
			dst.Set(x, y1+t, col)
			dst.Set(x, y2-t, col)
		}
		for y := y1; y <= y2; y++ {
			dst.Set(x1+t, y, col)
			dst.Set(x2-t, y, col)
		}
	}
}

const portalHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Autonomous Design Proofing Agent</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #eef2f5; margin: 0; padding: 24px; color: #2b2d42; }
    .container { max-width: 800px; margin: 0 auto; background: #ffffff; padding: 28px; border-radius: 12px; box-shadow: 0 4px 16px rgba(0,0,0,0.06); }
    h1 { margin-top: 0; color: #1d3557; font-size: 24px; }
    .upload-zone { border: 2px dashed #457b9d; border-radius: 8px; padding: 24px; text-align: center; background: #f8fafc; margin-bottom: 20px; }
    input[type="file"], input[type="number"] { margin: 8px 0; }
    .btn { padding: 10px 18px; border: none; border-radius: 6px; font-weight: 600; cursor: pointer; font-size: 15px; }
    .btn-primary { background: #1d3557; color: white; }
    .btn-approve { background: #2a9d8f; color: white; width: 48%; }
    .btn-reject { background: #e76f51; color: white; width: 48%; }
    .proof-card { margin-top: 24px; border: 1px solid #e2e8f0; border-radius: 8px; padding: 18px; display: none; }
    .proof-img { max-width: 100%; height: auto; border-radius: 4px; border: 1px solid #ccc; display: block; margin: 12px auto; }
    .badge { display: inline-block; padding: 4px 10px; border-radius: 12px; font-size: 13px; font-weight: 700; }
    .badge-ok { background: #d4edda; color: #155724; }
    .badge-warn { background: #fff3cd; color: #856404; }
    .status-alert { padding: 12px; border-radius: 6px; margin-top: 14px; font-weight: 600; text-align: center; }
  </style>
</head>
<body>
<div class="container">
  <h1>🎨 Autonomous Design Proofing Portal</h1>
  <p>Upload artwork to verify resolution, inspect cut lines, and create a digital proof.</p>

  <div class="upload-zone">
    <input type="file" id="artworkFile" accept="image/*"><br>
    <label>Target Width (Inches): <input type="number" id="wIn" value="4" step="0.1" style="width: 60px;"></label>
    <label>Target Height (Inches): <input type="number" id="hIn" value="6" step="0.1" style="width: 60px;"></label><br><br>
    <button class="btn btn-primary" onclick="uploadArtwork()">Analyze & Generate Proof</button>
  </div>

  <div id="proofCard" class="proof-card">
    <h2 id="proofTitle" style="margin-top:0;">Digital Proof</h2>
    <div id="stats"></div>
    <div id="defects" style="margin: 10px 0;"></div>

    <img id="proofPreview" class="proof-img" src="" alt="Proof Preview">

    <textarea id="notes" placeholder="Enter optional revision instructions or comments..." rows="3" style="width: 96%; margin: 10px 0; padding: 8px; border-radius: 6px; border: 1px solid #ccc;"></textarea>

    <div style="display: flex; justify-content: space-between; margin-top: 10px;">
      <button class="btn btn-approve" onclick="submitDecision('APPROVED')">✓ Approve Proof</button>
      <button class="btn btn-reject" onclick="submitDecision('CHANGES_REQUESTED')">✗ Request Changes</button>
    </div>

    <div id="decisionResult"></div>
  </div>
</div>

<script>
let currentProofId = '';

async function uploadArtwork() {
  const fileInput = document.getElementById('artworkFile');
  if (!fileInput.files[0]) {
    alert("Please select an image file first.");
    return;
  }

  const formData = new FormData();
  formData.append('artwork', fileInput.files[0]);
  formData.append('width_inches', document.getElementById('wIn').value);
  formData.append('height_inches', document.getElementById('hIn').value);

  const res = await fetch('/api/upload', { method: 'POST', body: formData });
  const data = await res.json();

  if (!res.ok) {
    alert(data.error || "Upload failed");
    return;
  }

  currentProofId = data.id;
  document.getElementById('proofCard').style.display = 'block';
  document.getElementById('proofTitle').innerText = 'Proof ' + data.id + ' (' + data.file_name + ')';
  document.getElementById('proofPreview').src = '/api/proof/' + data.id + '?t=' + Date.now();

  let dpiBadge = data.effective_dpi >= 300 ? '<span class="badge badge-ok">' + data.effective_dpi + ' DPI (High Quality)</span>' : '<span class="badge badge-warn">' + data.effective_dpi + ' DPI (Suboptimal)</span>';
  document.getElementById('stats').innerHTML = '<p><strong>Dimensions:</strong> ' + data.pixel_width + 'x' + data.pixel_height + ' px | <strong>Resolution:</strong> ' + dpiBadge + '</p>';

  let defectsHtml = '';
  if (data.defects && data.defects.length > 0) {
    data.defects.forEach(d => {
      defectsHtml += '<div class="status-alert badge-warn">' + d.severity + ': ' + d.message + '</div>';
    });
  } else {
    defectsHtml = '<div class="status-alert badge-ok">✓ No print defects found. Ready for printing.</div>';
  }
  document.getElementById('defects').innerHTML = defectsHtml;
}

async function submitDecision(decision) {
  if (!currentProofId) return;
  const notes = document.getElementById('notes').value;

  const res = await fetch('/api/decision', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id: currentProofId, decision: decision, notes: notes })
  });

  const data = await res.json();
  const alertColor = decision === 'APPROVED' ? '#d4edda' : '#f8d7da';
  const textColor = decision === 'APPROVED' ? '#155724' : '#721c24';

  document.getElementById('decisionResult').innerHTML = 
    '<div class="status-alert" style="background:' + alertColor + '; color:' + textColor + ';">Decision: ' + data.status + ' recorded at ' + new Date(data.decided_at).toLocaleTimeString() + '</div>';
}
</script>
</body>
</html>`
