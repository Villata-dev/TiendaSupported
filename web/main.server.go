package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"tiendasupported/modules" // Cambiado de "../modules"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	products     = make([]modules.Product, 0)
	users        = make([]modules.User, 0)
	sessions     = make([]modules.Session, 0)
	productIDSeq = 1
	userIDSeq    = 1
)

func main() {
	mux := http.NewServeMux()

	// Archivos estáticos
	fs := http.FileServer(http.Dir("web/public"))
	mux.Handle("/", fs)

	// API endpoints
	mux.HandleFunc("/api/v1/products", authMiddleware(productsHandler))
	mux.HandleFunc("/api/v1/products/", authMiddleware(productHandler)) // Note el "/" al final
	mux.HandleFunc("/api/auth/register", registerHandler)
	mux.HandleFunc("/api/auth/login", loginHandler)
	mux.HandleFunc("/api/auth/logout", logoutHandler)

	log.Println("Servidor iniciado en http://localhost:8080")
	log.Printf("Iniciando servidor con %d productos y %d usuarios", len(products), len(users))
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// Middleware de autenticación
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "No autorizado", http.StatusUnauthorized)
			return
		}

		// Buscar sesión válida
		var validSession *modules.Session
		for _, session := range sessions {
			if session.ID == cookie.Value && session.ExpiresAt.After(time.Now()) {
				validSession = &session
				break
			}
		}

		if validSession == nil {
			http.Error(w, "Sesión inválida", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// Handler de registro
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("Error decodificando registro: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verificar si el usuario ya existe
	for _, u := range users {
		if u.Username == credentials.Username {
			http.Error(w, "Usuario ya existe", http.StatusBadRequest)
			return
		}
	}

	// Hashear contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hasheando contraseña: %v", err)
		http.Error(w, "Error al procesar la contraseña", http.StatusInternalServerError)
		return
	}

	// Crear nuevo usuario
	newUser := modules.User{
		ID:           userIDSeq,
		Username:     credentials.Username,
		PasswordHash: string(hashedPassword),
		Role:         "user",
		CreatedAt:    time.Now(),
	}
	userIDSeq++
	users = append(users, newUser)

	log.Printf("Usuario registrado exitosamente: %s", newUser.Username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario registrado exitosamente"})
}

// Handler de login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("Error decodificando credenciales: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log para depuración
	log.Printf("Intento de login para usuario: %s", credentials.Username)
	log.Printf("Usuarios registrados: %+v", users)

	// Buscar usuario
	var user *modules.User
	for i := range users {
		if users[i].Username == credentials.Username {
			user = &users[i] // Importante: usar la referencia al elemento del slice
			break
		}
	}

	if user == nil {
		log.Printf("Usuario no encontrado: %s", credentials.Username)
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	// Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {
		log.Printf("Contraseña incorrecta para usuario: %s", credentials.Username)
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	// Crear sesión
	session := modules.Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	sessions = append(sessions, session)

	// Establecer cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Login exitoso",
		"userId":   strconv.Itoa(user.ID),
		"username": user.Username,
	})
}

// Handler para la colección de productos
func productsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Método %s en /api/v1/products", r.Method)

	switch r.Method {
	case http.MethodGet:
		log.Printf("Productos actuales: %v", products)
		json.NewEncoder(w).Encode(products)

	case http.MethodPost:
		var product modules.Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			log.Printf("Error decodificando producto: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Nuevo producto recibido: %v", product)
		product.ID = productIDSeq
		productIDSeq++
		product.CreatedAt = time.Now()
		product.UpdatedAt = time.Now()

		products = append(products, product)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(product)

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// Handler para producto individual
func productHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extraer ID del path
	idStr := r.URL.Path[len("/api/v1/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Buscar producto
	var productIndex = -1
	for i, p := range products {
		if p.ID == id {
			productIndex = i
			break
		}
	}

	if productIndex == -1 {
		http.Error(w, "Producto no encontrado", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(products[productIndex])

	case http.MethodPut:
		var updatedProduct modules.Product
		if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updatedProduct.ID = id
		updatedProduct.CreatedAt = products[productIndex].CreatedAt
		updatedProduct.UpdatedAt = time.Now()

		products[productIndex] = updatedProduct
		json.NewEncoder(w).Encode(updatedProduct)

	case http.MethodDelete:
		products = append(products[:productIndex], products[productIndex+1:]...)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Producto eliminado exitosamente"})

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// Handler de logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Eliminar cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout exitoso"})
}
