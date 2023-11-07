package controllers

import (
	"fmt"
	"log"
	"net/http"

	"projectfiber/middleware"
	"projectfiber/models"
	"projectfiber/repository"
	"projectfiber/usecase"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SignUpUser(c *fiber.Ctx) error {
	var request models.ReqSignUp

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		}) //jika terdapat error saat menarik isi db maka akan melempar error
	}

	// if request.Name == "" || request.Email == "" || request.Username == "" || request.Password == "" {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"message": "Harap mengisi semua form!",
	// 	}) //jika salah satu field belum terisi maka akan melempar pesan
	// }
	if request.Name == "" || request.Email == "" || request.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Harap mengisi semua form!",
		})
	}
	//cek apakah email sudah ada didatabase
	errEmail := usecase.CekEmail(request.Email)
	if errEmail != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": errEmail.Error(),
		})
	}
	//cek apakah name sudah ada didatabase
	errName := usecase.CekName(request.Name)
	if errName != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": errName.Error(),
		})
	}
	//cek apakah username sudah ada di database
	// errUname := usecase.CekUsername(request.Username)
	// if errUname != nil {
	// 	return c.Status(errUname.Code).JSON(fiber.Map{
	// 		"message": errUname.Message,
	// 	})
	// }
	//cek apakah password sudah ada didatabase
	errPass := usecase.CekPassword(request.Password)
	if errPass != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": errPass.Error(),
		})
	}
	//hashing password yang akan mengambil data password dari db. jika terdapat error maka aka melempar error
	hashedpassword, err := usecase.HashPassword(request.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	// jika username belum ada di database, maka buat akun baru
	newUser := models.Users{
		Name:  request.Name,
		Email: request.Email,
		//Username: request.Username,
		Password: hashedpassword,
	}
	if err := repository.DB.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		}) //jika error dalam db maka akan melempar error
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Anda sukses mendaftar.",
	}) //jika tidak maka akan sukses mendaftar
}

func LoginUser(c *fiber.Ctx) error {
	var loginPayload models.LoginPayLoad
	var users models.Users

	if err := c.BodyParser(&loginPayload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		}) //jika terdapat error saat menarik isi db maka akan melempar error
	}
	//user := loginPayload.Username
	email := loginPayload.Email
	pass := loginPayload.Password
	if err := repository.DB.First(&users, "email = ?", email).Error; err != nil {
		log.Print("Data tidak ditemukan.")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Data tidak ditemukan.",
		})
	}
	// if err := repository.DB.First(&users, "username", user).Error; err != nil {
	// 	log.Print("Data tidak ditemukan.")
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"message": "Data tidak ditemukan",
	// 	}) //jika username tidak ditemukan atau tidak sesuai dengan database maka akan melempar error
	// }

	if !users.AccStatus {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Data tidak ditemukan",
		})
	}

	var hashedpassword string
	if err := repository.DB.Select("password").First(&users, "email", email).Scan(&hashedpassword).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Data tidak ditemukan",
		})
	}
	// if err := repository.DB.Select("password").First(&users, "username", user).Scan(&hashedpassword).Error; err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"message": "Data tidak ditemukan!",
	// 	})
	// }
	match := usecase.CheckPasswordHash(pass, hashedpassword)
	if match {
		emailID := users.ID.String()
		acctoken, _, err := middleware.CreateToken(true, emailID)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Gagal membuat access token",
			}) //jika terdapat error pada middleware saat membuat access token
		}
		tref, tref_acc, err := middleware.CreateToken(false, emailID)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Gagal membuat refresh token",
			}) //jika terdapat error pada middleware saat membuat refresh token
		}
		_, err = CreateSession(emailID, tref, tref_acc, c)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Tidak bisa membuat session!",
			}) //jika saat membuat session dengan user id, acc token dan ref token gagal
		}
		//update accestoken dan refresh token
		if repository.DB.Model(&users).Where("email", email).Updates(models.Users{AccessToken: acctoken, RefreshToken: tref}).RowsAffected == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Tidak dapat mengupdate access token!",
			})
		}
		// if repository.DB.Model(&users).Where("username", user).Updates(models.Users{AccessToken: acctoken, RefreshToken: tref}).RowsAffected == 0 {
		// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 		"message": "Tidak dapat mengupdate access token!",
		// 	})
		// }
		return c.JSON(users)
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Password tidak cocok!",
		})
	}
}

func CreateSession(emailID string, refToken string, payloadClaims jwt.MapClaims, c *fiber.Ctx) (*models.SessionUser, error) {
	expUnixTimestamp := payloadClaims["exp"].(int64)
	expTime := time.Unix(expUnixTimestamp, 0)

	sessionIDUUID, err := uuid.Parse(payloadClaims["sessionID"].(string))
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	emailIDUUID, err := uuid.Parse(payloadClaims["emailID"].(string))
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	var session models.SessionUser
	session.ID = sessionIDUUID
	session.EmailID = emailIDUUID
	session.RefreshToken = refToken
	session.UserAgent = c.Get("User-Agent")
	session.ClientIp = c.IP()
	session.ExpiresAt = expTime

	if err := repository.DB.Create(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func LogOutUser(c *fiber.Ctx) error {
	// emailID, err := middleware.GetEmailIdByClaims(c)
	// if err != nil {
	// 	return err
	// }
	// email, err := repository.GetUserByID(emailID)
	// if err != nil {
	// 	return err
	// }
	// claims, err := middleware.GetClaimsByTokenString(c, email.RefreshToken)
	// if err != nil {
	// 	return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
	// 		"message": "Invalid claims!",
	// 	})
	// }
	// sessionID, ok := claims["sessionID"].(string)
	// if !ok || sessionID == "" {
	// 	// respons error jika sessionID kosong
	// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
	// 		"message": "Database tidak dapat menemukan sessionID.",
	// 	})
	// }
	// err = middleware.DeleteSession(sessionID)
	// if err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"message": "Tidak dapat menemukan session id! Silahkan login kembali!",
	// 	})
	// }
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Anda telah berhasil LogOut",
	})
}

func GetLoginUsername(c *fiber.Ctx) error {
	emails := c.Params("email")
	var users models.Users
	if err := repository.DB.Where("email = ?", emails).First(&users).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "Data tidak ditemukan",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Unexpected Error",
		})
	}
	return c.JSON(users)
}

type JobUserController struct {
	DB *gorm.DB
}

func NewJobUserController(db *gorm.DB) *JobUserController {
	return &JobUserController{DB: db}
}

func (controllers *JobUserController) UploadDocument(c *fiber.Ctx) error {
	file_, err := c.FormFile("document")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal mendapatkan dokumen",
		})
	}
	result, err := usecase.PostDocumentUser(file_)
	if err != nil {
		return err
	}
	// emailID, err := middleware.GetEmailIdByClaims(c)
	// if err != nil {
	// 	return err
	// }
	newJobUser := models.JobUser{
		//EmailID: uuid.MustParse(emailID),
		Job: result.HasilPrediksi,
	}
	if err := controllers.DB.Create(&newJobUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Dokumen anda gagal disimpan",
		})
	}
	fmt.Print(result)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Berdasarkan CV anda, hasil pekerjaan yang cocok dengan anda adalah: " + result.HasilPrediksi,
	})
}
