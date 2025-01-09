package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Medication struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Dosage string    `json:"dosage"`
}

type Server struct {
	medications map[string]*Medication
	mu          sync.Mutex
}

func NewServer() *Server {
	return &Server{
		medications: make(map[string]*Medication),
	}
}

func (s *Server) addMedication(w http.ResponseWriter, r *http.Request) {
	var med Medication
	if err := json.NewDecoder(r.Body).Decode(&med); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.medications[med.Name]; exists {
		http.Error(w, "Medication already exists", http.StatusConflict)
		return
	}

	s.medications[med.Name] = &med
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) getMedication(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing medication name", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	med, exists := s.medications[name]
	if !exists {
		http.Error(w, "Medication not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(med)
}

func (s *Server) getAllMedications(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	medications := make([]Medication, 0, len(s.medications))
	for _, med := range s.medications {
		medications = append(medications, *med)
	}

	json.NewEncoder(w).Encode(medications)
}

func (s *Server) updateMedicationCount(w http.ResponseWriter, r *http.Request) {
	var med Medication
	if err := json.NewDecoder(r.Body).Decode(&med); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existingMed, exists := s.medications[med.Name]
	if !exists {
		http.Error(w, "Medication not found", http.StatusNotFound)
		return
	}

	existingMed.Count = med.Count
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteMedication(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing medication name", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.medications[name]; !exists {
		http.Error(w, "Medication not found", http.StatusNotFound)
		return
	}

	delete(s.medications, name)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) routes() {
	http.HandleFunc("/add", s.addMedication)
	http.HandleFunc("/get", s.getMedication)
	http.HandleFunc("/getAll", s.getAllMedications)
	http.HandleFunc("/update", s.updateMedicationCount)
	http.HandleFunc("/delete", s.deleteMedication)
}

func main() {
	server := NewServer()
	server.routes()

	log.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
