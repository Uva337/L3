package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/wb-go/wbf/ginext"
)

// Константы для ролей
const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleViewer  = "viewer"
)

func parseFakeToken(authHeader string) (string, string, error) {
	if authHeader == "" {
		return "", "", errors.New("отсутствует заголовок Authorization")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", "", errors.New("неверный формат токена")
	}

	tokenData := strings.Split(parts[1], ":")
	if len(tokenData) != 2 {
		return "", "", errors.New("неверный формат данных внутри токена")
	}

	return tokenData[0], tokenData[1], nil
}

// AuthMiddleware проверяет наличие прав доступа
func AuthMiddleware(allowedRoles ...string) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		authHeader := c.GetHeader("Authorization")

		username, role, err := parseFakeToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, ginext.H{"error": err.Error()})
			c.Abort()
			return
		}

		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.JSON(http.StatusForbidden, ginext.H{"error": "недостаточно прав для выполнения операции"})
			c.Abort()
			return
		}

		c.Set("username", username)
		c.Set("role", role)

		c.Next()
	}
}
