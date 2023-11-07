package middleware

import (
	"errors"
	"fmt"
	"os"
	"projectfiber/models"
	"projectfiber/repository"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenExpiration  = 5 * time.Minute
	refreshTokenExpiration = 7 * 24 * time.Hour
)

func AuthUserLogin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := os.Getenv("PRIVATE_KEY")
		authHeader := c.Get("Authorization")
		if len(authHeader) == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Silakan sign-in terlebih dahulu",
			}) //jika panjang dari isian kotak authorization pada postman masih kosong, maka akan muncul error
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Silakan sign-in ulang",
			}) //jika isi dari token yang dimasukkan tidak ada
		}
		claimvalid, err := GetClaimsByTokenString(c, tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Silakan sign-up atau sign-in terlebih dahulu",
			}) //claimvalid mempunyai 2 kondisi yaitu valid dan error, jika error saat pengecekan maka akan error
		}
		if claimvalid["ticket"] == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Silakan sign-up atau sign-in terlebih dahulu.",
			}) //jika claimvalidnya itu berupa tiket access token dan kosong, maka akan error
		}
		isTicket := claimvalid["ticket"].(bool)
		if !isTicket {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Silakan sign-up atau sign-in terlebih dahulu.",
			}) //jika pas dicek ternyata tiket nya itu tidak ada maka akan err
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		}) //
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Akses tidak diberikan",
			})
		}
		c.Locals("emailID", token)
		return c.Next()
	}
}

func GetClaimsByTokenString(c *fiber.Ctx, token string) (jwt.MapClaims, error) {
	secret := os.Getenv("PRIVATE_KEY") // secret key untuk signing JWT
	// verifikasi refresh-token JWT
	tokenString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// memeriksa tipe metode signing
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// mengembalikan secret key untuk verifikasi
		return []byte(secret), nil
	})
	//jika error nya ada atau tokenstring nya tidak valid
	if err != nil || !tokenString.Valid {
		return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Akses tidak diberikan!",
		})
	}
	claims, ok := tokenString.Claims.(jwt.MapClaims)
	//jika error nya ada
	if !ok {
		return nil, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Token tidak ditemukan",
		})
	}
	return claims, nil
}

func RenewAccToken(c *fiber.Ctx) error {
	var req models.RefreshTokenReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Tidak dapat melanjutkan proses",
		}) //jika terdapat err pada saat pengambilan data dari db
	}
	claimsPayload, err := GetClaimsByTokenString(c, req.RefreshToken)
	if err != nil {
		return err
	}
	sessionID, ok := claimsPayload["sessionID"].(string)
	if !ok || sessionID == "" {
		// respons error jika sessionID kosong
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "SessionID tidak ditemukan",
		})
	}
	session, err := GetSession(sessionID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Tidak dapat menemukan session id!",
		})
	}
	if *session.IsBlocked {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Session ini diblokir",
		})
	}
	emailID := claimsPayload["emailID"].(string)
	//user, err := repository.GetUserByID(userID)
	email, err := repository.GetUserByID(emailID)
	if err != nil {
		return err
	}
	if session.EmailID != email.ID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Incorrect session user!",
		})
	}
	if session.RefreshToken != email.RefreshToken {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Session token tidak cocok",
		})
	}
	if time.Now().After(session.ExpiresAt) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Sesi anda telah habis, silakan sign-in kembali",
		})
	}
	secret := os.Getenv("PRIVATE_KEY") // Secret key untuk signing JWT
	// verifikasi refresh-token JWT
	refreshToken, err := jwt.Parse(session.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		// Memeriksa tipe metode signing
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// secret key untuk verifikasi
		return []byte(secret), nil
	})
	if err != nil || !refreshToken.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Akses tidak diberikan",
		})
	}

	// jika refresh token valid, buat access token baru dengan waktu kadaluarsa yang baru (misalnya, 15 menit)
	newToken, _, err := CreateToken(true, emailID)
	if err != nil {
		return err
	}
	//update accestoken
	if repository.DB.Model(&email).Update("access_token", newToken).RowsAffected == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Tidak dapat mengupdate access token!",
		})
	}
	resp := models.RefreshTokenRes{
		AccessToken: newToken,
	}
	// menyimpan token dalam konteks untuk digunakan oleh handler selanjutnya
	c.Locals("emailID")
	// melanjutkan eksekusi ke handler selanjutnya
	return c.Status(fiber.StatusOK).JSON(resp)
}

func GetSession(sessionID string) (*models.SessionUser, error) {
	var data models.SessionUser
	if err := repository.DB.First(&data, "id", sessionID).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func DeleteSession(sessionID string) error {
	var data models.SessionUser
	if repository.DB.Delete(&data, "id", sessionID).RowsAffected == 0 {
		return fmt.Errorf("Session ID tidak ditemukan")
	}
	return nil
}

func GetClaims(c *fiber.Ctx) (jwt.MapClaims, error) {
	tokenInterface := c.Locals("emailID")

	if tokenInterface == nil {
		return nil, errors.New("Tidak ada token ditemukan.")
	}

	token, ok := tokenInterface.(*jwt.Token)
	if !ok || token == nil {
		return nil, errors.New("Invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Invalid token claims")
	}

	return claims, nil
}

func CreateToken(isAccessToken bool, emailID string) (string, jwt.MapClaims, error) {
	// Menentukan waktu kedaluwarsa token berdasarkan jenisnya
	secretKey := os.Getenv("PRIVATE_KEY")
	var expiration time.Time
	if isAccessToken {
		expiration = time.Now().Add(accessTokenExpiration)
	} else {
		expiration = time.Now().Add(refreshTokenExpiration)
	}
	//membuat claims
	claims := jwt.MapClaims{
		"emailID": emailID,
		"exp":     expiration.Unix(), // kadaluwarsa setelah 24 jam
		"iat":     expiration.Unix(),
	}
	// jika isAccessToken adalah true, tambahkan sessionID ke dalam claims
	if !isAccessToken {
		claims["sessionID"] = uuid.NewString()
	}
	claims["ticket"] = isAccessToken
	//membuat token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// menghasilkan string token menggunakan secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", claims, err
	}
	return tokenString, claims, nil
}

func GetEmailIdByClaims(c *fiber.Ctx) (string, error) {
	claims, err := GetClaims(c)
	if err != nil {
		return "", err
	}
	emailID := claims["emailID"].(string)
	return emailID, nil
}
