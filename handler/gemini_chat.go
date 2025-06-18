package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

func GeminiChatPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/gemini_chat.html")
}

func GeminiChatAPI(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		http.Error(w, "GOOGLE_API_KEY manquante", http.StatusInternalServerError)
		return
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": req.Message},
				},
			},
		},
		"generation_config": map[string]interface{}{
			"temperature":       0.7,
			"max_output_tokens": 256,
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Erreur JSON interne", http.StatusInternalServerError)
		return
	}

	url := "https://generativelanguage.googleapis.com/v1/models/gemini-1.5-flash:generateContent?key=" + apiKey
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Println("Erreur réseau Gemini :", err)
		http.Error(w, "Erreur réseau Gemini", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Gemini %s\n%s\n", resp.Status, string(body))
		http.Error(w, "Gemini a renvoyé "+resp.Status+": "+string(body), http.StatusBadGateway)
		return
	}

	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		http.Error(w, "Réponse Gemini invalide", http.StatusInternalServerError)
		return
	}

	reply := ""
	if len(out.Candidates) > 0 && len(out.Candidates[0].Content.Parts) > 0 {
		reply = out.Candidates[0].Content.Parts[0].Text
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Reply: reply})
}
