# Evaluación 3 - API RESTful Completa y Cliente Web

## Descripción General
Sistema de gestión de productos con autenticación y roles de usuario. La API proporciona endpoints para gestionar productos y autenticación, mientras que el cliente web ofrece una interfaz interactiva usando Web Components.

## Endpoints CRUD

### Productos

| Método | Ruta | Descripción | Params | Body | Ejemplo Petición | Respuesta Éxito | Errores |
|--------|------|-------------|---------|------|-----------------|-----------------|----------|
| GET | `/api/v1/products` | Obtener lista | `?page=1&limit=10` | - | `GET /api/v1/products` | `[{"id": 1, "name": "Producto", ...}]` | 401, 500 |
| GET | `/api/v1/products/{id}` | Obtener uno | `id` | - | `GET /api/v1/products/1` | `{"id": 1, "name": "Producto", ...}` | 401, 404 |
| POST | `/api/v1/products` | Crear nuevo | - | `{"name": "", "price": 0}` | `POST /api/v1/products` | `{"id": 1, ...}` | 400, 401, 403 |
| PUT | `/api/v1/products/{id}` | Actualizar | `id` | `{"name": "", "price": 0}` | `PUT /api/v1/products/1` | `{"id": 1, ...}` | 400, 401, 403, 404 |
| DELETE | `/api/v1/products/{id}` | Eliminar | `id` | - | `DELETE /api/v1/products/1` | `{"message": "ok"}` | 401, 403, 404 |

### Autenticación

| Método | Ruta | Descripción | Body | Cookies | Ejemplo | Respuesta |
|--------|------|-------------|------|----------|----------|------------|
| POST | `/api/auth/register` | Registro | `{"username": "", "password": ""}` | - | `POST /api/auth/register` | `{"message": "ok"}` |
| POST | `/api/auth/login` | Login | `{"username": "", "password": ""}` | Set-Cookie | `POST /api/auth/login` | `{"token": "..."}` |
| POST | `/api/auth/logout` | Logout | - | Clear-Cookie | `POST /api/auth/logout` | `{"message": "ok"}` |

## Middleware y Permisos

### Sistema de Autenticación
- Middleware verifica token en cookie para rutas protegidas
- Sin token válido retorna 401 Unauthorized
- Token incluye ID de usuario y rol

### Roles y Permisos
- **Admin**: CRUD completo
- **Usuario**: Solo lectura
- Acciones no permitidas retornan 403 Forbidden

## Cómo Ejecutar el Servidor

1. Navegar al directorio del proyecto:
```bash
cd c:\Users\m\TiendaSupported
```

2. Compilar y ejecutar:
```bash
go build web/main.server.go
./main.server.exe
```

O ejecutar directamente:
```bash
go run web/main.server.go
```

El servidor iniciará en http://localhost:8080

## Cómo Probar (Cliente Web)

1. **Acceder al Cliente Web**
   - Abrir http://localhost:8080 en el navegador

2. **Autenticación**
   - Iniciar sesión como admin:
     - Usuario: admin
     - Contraseña: admin123
   - O como usuario normal:
     - Usuario: user
     - Contraseña: user123

3. **Gestión de Productos**
   - Ver lista de productos
   - Usar filtros y búsqueda
   - Crear nuevo producto (solo admin)
   - Editar producto (solo admin)
   - Eliminar producto (solo admin)

4. **Navegación**
   - Usar paginación
   - Ajustar items por página
   - Ordenar por diferentes campos

## Código Relevante

### Backend (Go)

#### Middleware de Autenticación
```go
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_token")
        if err != nil {
            http.Error(w, "No autorizado", http.StatusUnauthorized)
            return
        }
        // Validar sesión y permisos
        ctx := context.WithValue(r.Context(), userContextKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}
```

#### Handler de Productos
```go
func productsHandler(w http.ResponseWriter, r *http.Request) {
    user := r.Context().Value(userContextKey).(*models.User)
    switch r.Method {
    case http.MethodGet:
        json.NewEncoder(w).Encode(products)
    case http.MethodPost:
        if user.Role != "Admin" {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        // Crear producto
    }
}
```

### Frontend (JavaScript)

#### Web Component EditableList
```javascript
class EditableList extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
    }

    static get styles() {
        return `
            :host {
                display: block;
                font-family: system-ui;
            }
            table {
                width: 100%;
                border-collapse: collapse;
            }
        `;
    }

    async loadProducts() {
        const response = await fetch('/api/v1/products');
        const products = await response.json();
        this.renderProducts(products);
    }
}

customElements.define('editable-list', EditableList);
```

#### Manejo de Autenticación
```javascript
async function login(credentials) {
    const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(credentials)
    });
    
    if (response.ok) {
        const data = await response.json();
        showMessage('Login exitoso');
        loadProducts();
    }
}
```
