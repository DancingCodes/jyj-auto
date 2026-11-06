package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/api/info", getInfo)
	mux.HandleFunc("/api/machine-export", machineExport)
	return mux
}

func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "请求方式错误")
		return
	}
	_, _ = w.Write([]byte("ok"))
}

func getInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "请求方式错误")
		return
	}
	result, err := GetInfo(r.Context())
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeJSON(w, result)
}

func machineExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "请求方式错误")
		return
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		writeError(w, "date 格式应为 YYYY-MM-DD")
		return
	}
	file, err := BuildMachineExport(r.Context(), date)
	if err != nil {
		writeError(w, err.Error())
		return
	}
	fileName := time.Now().Format("200601021504") + "-机具查询导出.xlsx"
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	_, _ = w.Write(file)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(message))
}
