package models

import "time"

// Product representa un producto en la tienda
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// User representa un usuario del sistema
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Ocultar el hash de la contraseña en JSON
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

// SessionID es un tipo para el ID de sesión (UUID)
type SessionID string

// Session representa una sesión de usuario activa
type Session struct {
	ID        SessionID `json:"id"`
	UserID    int       `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}
