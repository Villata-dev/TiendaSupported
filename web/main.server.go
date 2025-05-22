package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings" // Importar para strings.TrimSpace
	"time"

	models "TiendaSupported/modules" // ¡IMPORTACIÓN CORREGIDA para el nuevo nombre del módulo!

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Definir un tipo de clave de contexto personalizado para evitar colisiones
type contextKey string

// Declarar una constante para la clave del usuario en el contexto
const userContextKey contextKey = "user"

var (
	// Usamos slices para almacenar en memoria, inicializados con datos de prueba
	products     = make([]models.Product, 0)
	users        = make([]models.User, 0)
	sessions     = make([]models.Session, 0)
	productIDSeq = 1
	userIDSeq    = 1
)

func main() {
	mux := http.NewServeMux()

	// Inicializar datos de prueba al inicio del servidor
	initializeData()

	// --- Manejo de Archivos Estáticos y Rutas de la API ---
	// ¡CORRECCIÓN CLAVE! Servir archivos estáticos bajo un prefijo /static/
	// y manejar la ruta raíz explícitamente para index.html.
	// Esto evita que el FileServer capture las rutas de la API.
	fs := http.FileServer(http.Dir("web/public"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Manejar la ruta raíz "/" para servir index.html (SPA)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Asegurarse de que solo se sirva index.html para la raíz y no para otras rutas no API
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/public/index.html")
	})

	// API endpoints
	// Orden de las rutas: Las rutas exactas primero, luego las rutas con parámetros.
	// Esto ayuda a evitar que "/api/v1/products/{id}" capture "/api/v1/products"
	mux.HandleFunc("/api/v1/products", authMiddleware(productsHandler))
	// Nota: El patrón "/api/v1/products/" con la barra final es para capturar paths como "/api/v1/products/123"
	mux.HandleFunc("/api/v1/products/", authMiddleware(productHandler))

	mux.HandleFunc("/api/auth/register", registerHandler)
	mux.HandleFunc("/api/auth/login", loginHandler)
	mux.HandleFunc("/api/auth/logout", logoutHandler)
	mux.HandleFunc("/api/auth/check-session", checkSessionHandler) // Nueva ruta para verificar sesión

	log.Println("Servidor iniciado en http://localhost:8080")
	log.Printf("Iniciando servidor con %d productos y %d usuarios", len(products), len(users))
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// initializeData crea algunos productos y usuarios de prueba
func initializeData() {
	log.Println("⏳ Inicializando datos de ejemplo...")

	// Crear productos de ejemplo
	products = append(products, models.Product{
		ID:          productIDSeq,
		Name:        "Laptop Gamer Pro",
		Description: "Potente laptop para juegos de última generación con RTX 4090",
		Price:       1850.75,
		Stock:       8,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	productIDSeq++
	products = append(products, models.Product{
		ID:          productIDSeq,
		Name:        "Teclado Mecánico RGB HyperX",
		Description: "Teclado con switches Cherry MX Red y retroiluminación RGB personalizable",
		Price:       110.00,
		Stock:       45,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	productIDSeq++
	products = append(products, models.Product{
		ID:          productIDSeq,
		Name:        "Monitor Curvo UltraWide 34\"",
		Description: "Monitor 4K de alta resolución para diseño y gaming inmersivo",
		Price:       499.99,
		Stock:       12,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	productIDSeq++
	log.Printf("✅ Inicializados %d productos de ejemplo.", len(products))

	// Crear usuarios de prueba
	registerTestUser := func(username, password, role string) {
		for _, u := range users {
			if u.Username == username {
				log.Printf("ℹ️ Usuario de prueba '%s' (Rol: %s) ya existe.", username, role)
				return
			}
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("❌ Fatal: No se pudo hashear contraseña para %s: %v", username, err)
		}
		newUser := models.User{
			ID:           userIDSeq,
			Username:     username,
			PasswordHash: string(hashedPassword),
			Role:         role,
			CreatedAt:    time.Now(),
		}
		userIDSeq++
		users = append(users, newUser)
		log.Printf("✅ Usuario de prueba '%s' (Rol: %s) registrado.", username, role)
	}

	registerTestUser("admin", "admin123", "Admin")    // Rol Admin
	registerTestUser("editor", "editor123", "Editor") // Rol Editor
	registerTestUser("user", "user123", "User")       // Rol Usuario normal
	log.Printf("✅ Inicialización de usuarios de prueba completada.")
}

// Middleware de autenticación
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Configurar CORS
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Añadir Authorization si se usa

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		log.Printf("Verificando autenticación para: %s %s", r.Method, r.URL.Path)

		// Verificar cookie de sesión
		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("Cookie 'session_token' no encontrada: %v", err)
			http.Error(w, "No autorizado: Cookie de sesión no encontrada", http.StatusUnauthorized)
			return
		}

		log.Printf("Cookie encontrada: %s", cookie.Value)

		// Buscar sesión válida
		var validSession *models.Session
		for i := range sessions { // Usar range con índice para obtener referencia modificable si fuera necesario
			session := &sessions[i] // Obtener la dirección de la sesión
			if session.ID == models.SessionID(cookie.Value) {
				if session.ExpiresAt.After(time.Now()) {
					validSession = session
					log.Printf("Sesión válida encontrada para usuario ID: %d", session.UserID)
					break
				} else {
					log.Printf("Sesión expirada para usuario ID: %d. Eliminando sesión.", session.UserID)
					// Eliminar sesión expirada del slice
					sessions = append(sessions[:i], sessions[i+1:]...)
					http.Error(w, "Sesión expirada", http.StatusUnauthorized)
					return
				}
			}
		}

		if validSession == nil {
			log.Printf("No se encontró sesión válida para el token: %s", cookie.Value)
			http.Error(w, "Sesión inválida", http.StatusUnauthorized)
			return
		}

		// Añadir información de usuario al contexto usando la clave personalizada
		var authenticatedUser *models.User
		for i := range users {
			if users[i].ID == validSession.UserID {
				authenticatedUser = &users[i]
				break
			}
		}

		if authenticatedUser == nil {
			log.Printf("Error interno: Usuario ID %d no encontrado para sesión válida.", validSession.UserID)
			http.Error(w, "Error interno de autenticación: Usuario no encontrado", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, authenticatedUser) // USANDO LA CLAVE PERSONALIZADA
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Handler de registro
func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Configurar CORS para este handler específico (o usar un wrapper global)
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(credentials.Username) == "" || strings.TrimSpace(credentials.Password) == "" {
		http.Error(w, "Nombre de usuario y contraseña no pueden estar vacíos", http.StatusBadRequest)
		return
	}

	// Verificar si el usuario ya existe
	for _, u := range users {
		if u.Username == credentials.Username {
			http.Error(w, "Usuario ya existe", http.StatusConflict) // 409 Conflict
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
	newUser := models.User{
		ID:           userIDSeq,
		Username:     credentials.Username,
		PasswordHash: string(hashedPassword),
		Role:         "user", // Rol por defecto
		CreatedAt:    time.Now(),
	}
	userIDSeq++
	users = append(users, newUser)

	log.Printf("Usuario registrado exitosamente: %s (ID: %d)", newUser.Username, newUser.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario registrado exitosamente"})
}

// Handler de login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Configurar CORS
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Intento de login para usuario: %s", credentials.Username)

	// Buscar usuario
	var user *models.User
	for i := range users {
		if users[i].Username == credentials.Username {
			user = &users[i]
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

	// Crear nueva sesión
	session := models.Session{
		ID:        models.SessionID(uuid.New().String()),
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	sessions = append(sessions, session)

	// Establecer cookie con configuración correcta
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    string(session.ID),
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Cambiar a 'true' en producción con HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(24 * time.Hour / time.Second), // 24 horas en segundos
	})

	// Responder con JSON incluyendo información del usuario
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Login exitoso",
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
	log.Printf("Login exitoso para usuario: %s", user.Username)
}

// Handler para la colección de productos
func productsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Método %s en /api/v1/products", r.Method)

	// Recuperar el usuario del contexto
	user, ok := r.Context().Value(userContextKey).(*models.User) // USANDO LA CLAVE PERSONALIZADA
	if !ok || user == nil {
		log.Printf("Error: Usuario no encontrado en el contexto para productsHandler.")
		http.Error(w, "Error interno de autenticación", http.StatusInternalServerError)
		return
	}
	log.Printf("productsHandler accedido por usuario: %s (Rol: %s)", user.Username, user.Role)

	switch r.Method {
	case http.MethodGet:
		// Asegurarse de que el slice de productos no sea nil si está vacío
		if products == nil {
			products = []models.Product{}
		}
		log.Printf("Productos actuales: %v", products)
		json.NewEncoder(w).Encode(products)

	case http.MethodPost:
		// Solo permitir POST si el usuario es Admin o Editor
		if user.Role != "Admin" && user.Role != "Editor" {
			http.Error(w, "Acceso denegado: No tienes permisos para agregar productos.", http.StatusForbidden)
			return
		}

		var product models.Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			log.Printf("Error decodificando producto: %v", err)
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if strings.TrimSpace(product.Name) == "" {
			http.Error(w, "El nombre del producto no puede estar vacío", http.StatusBadRequest)
			return
		}
		if product.Price < 0 {
			http.Error(w, "El precio del producto no puede ser negativo", http.StatusBadRequest)
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

	// Recuperar el usuario del contexto
	user, ok := r.Context().Value(userContextKey).(*models.User) // USANDO LA CLAVE PERSONALIZADA
	if !ok || user == nil {
		log.Printf("Error: Usuario no encontrado en el contexto para productHandler.")
		http.Error(w, "Error interno de autenticación", http.StatusInternalServerError)
		return
	}
	log.Printf("productHandler accedido por usuario: %s (Rol: %s)", user.Username, user.Role)

	// Extraer ID del path (ej: /api/v1/products/123 -> "123")
	// Usar strings.TrimPrefix para manejar el caso de la ruta base "/api/v1/products/"
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("ID inválido en la ruta: %s, error: %v", idStr, err)
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
		// Solo permitir PUT si el usuario es Admin o Editor
		if user.Role != "Admin" && user.Role != "Editor" {
			http.Error(w, "Acceso denegado: No tienes permisos para editar productos.", http.StatusForbidden)
			return
		}

		var updatedProduct models.Product
		if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
			log.Printf("Error decodificando producto para actualizar: %v", err)
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if strings.TrimSpace(updatedProduct.Name) == "" {
			http.Error(w, "El nombre del producto no puede estar vacío", http.StatusBadRequest)
			return
		}
		if updatedProduct.Price < 0 {
			http.Error(w, "El precio del producto no puede ser negativo", http.StatusBadRequest)
			return
		}

		updatedProduct.ID = id
		updatedProduct.CreatedAt = products[productIndex].CreatedAt // Mantener la fecha de creación original
		updatedProduct.UpdatedAt = time.Now()

		products[productIndex] = updatedProduct
		json.NewEncoder(w).Encode(updatedProduct)

	case http.MethodDelete:
		// Solo permitir DELETE si el usuario es Admin
		if user.Role != "Admin" {
			http.Error(w, "Acceso denegado: No tienes permisos para eliminar productos.", http.StatusForbidden)
			return
		}

		// Eliminar el producto del slice
		products = append(products[:productIndex], products[productIndex+1:]...)
		w.WriteHeader(http.StatusOK) // 200 OK para éxito de eliminación
		json.NewEncoder(w).Encode(map[string]string{"message": "Producto eliminado exitosamente"})

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// Handler de logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Configurar CORS
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Invalidar la cookie de sesión
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,                          // Cambiar a 'true' en producción con HTTPS
		Expires:  time.Now().Add(-1 * time.Hour), // Expira la cookie inmediatamente
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout exitoso"})
	log.Println("Sesión cerrada exitosamente.")
}

// Handler para verificar sesión
func checkSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Configurar CORS
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("Verificando sesión en /api/auth/check-session")

	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("Error al obtener cookie en check-session: %v", err)
		http.Error(w, "No autenticado: Cookie de sesión no encontrada", http.StatusUnauthorized)
		return
	}

	// Buscar sesión válida
	var validSession *models.Session
	for i := range sessions {
		session := &sessions[i]
		if session.ID == models.SessionID(cookie.Value) && session.ExpiresAt.After(time.Now()) {
			validSession = session
			break
		}
	}

	if validSession == nil {
		log.Printf("Sesión no válida o expirada en check-session para token: %s", cookie.Value)
		http.Error(w, "Sesión inválida o expirada", http.StatusUnauthorized)
		return
	}

	// Buscar el usuario asociado a la sesión
	var user *models.User
	for i := range users {
		if users[i].ID == validSession.UserID {
			user = &users[i]
			break
		}
	}

	if user == nil {
		log.Printf("Error interno: Usuario ID %d no encontrado para sesión válida.", validSession.UserID)
		http.Error(w, "Error interno: Usuario no encontrado", http.StatusInternalServerError)
		return
	}

	log.Printf("Sesión válida encontrada para usuario: %s (ID: %d, Rol: %s)", user.Username, user.ID, user.Role)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Sesión válida",
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}
