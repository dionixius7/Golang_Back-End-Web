package usecase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"mime/multipart"
	"net/http"

	"projectfiber/models"
	"projectfiber/repository"
	"regexp"

	"github.com/gofiber/fiber/v2"
)

func CekEmail(email string) error {
	var users models.Users
	var count int64
	if err := repository.DB.Model(&users).Where("email=?", email).Count(&count).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}
	if count != 0 {
		return fmt.Errorf("Email anda sudah terdaftar.", email)
	}
	return nil
}

func CekName(name string) error {
	var users models.Users
	var count int64
	if err := repository.DB.Model(&users).Where("name=?", name).Count(&count).Error; err != nil {
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
	}
	if count != 0 {
		return fmt.Errorf("Nama anda belum dimasukkan.")
	}
	return nil
}

// func CekUsername(username string) *fiber.Error {
// 	var users models.Users
// 	var count int64
// 	if err := repository.DB.Model(&users).Where("username=?", username).Count(&count).Error; err != nil {
// 		// Jika username sudah ada di database, maka return error
// 		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: err.Error()}
// 	}
// 	if count != 0 {
// 		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Username tidak boleh sama."}
// 	}
// 	return nil
// }

func CekPassword(password string) error {
	const (
		minLength      = 8
		minUppercase   = 1
		minDigits      = 1
		minSpecialChar = 1
	)

	if len(password) < minLength {
		return errors.New("Password kurang panjang!")
	}
	uppercaseCount := len(regexp.MustCompile(`[^A-Z]`).ReplaceAllString(password, ""))
	if uppercaseCount < minUppercase {
		return errors.New("Masukkan minimal 1 huruf kapital pada password!")
	}
	digitCount := len(regexp.MustCompile(`[^0-9]`).ReplaceAllString(password, ""))
	if digitCount < minDigits {
		return errors.New("Masukkan minimal 1 angka pada password!")
	}
	specialCharCount := len(regexp.MustCompile(`[a-zA-Z0-9]`).ReplaceAllString(password, ""))
	if specialCharCount < minSpecialChar {
		return errors.New("Masukkan minimal 1 karakter spesial pada password!")
	}
	return nil
}

func SendHttpRequest(req *http.Request, output interface{}) (*http.Response, error) {
	// Create an HTTP client
	client := &http.Client{}
	// Send the request
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return res, err
	}
	// Make sure to close the response body when done
	defer res.Body.Close()
	// Check if the response content type is "application/json"
	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf("error: unexpected content-type: %s", contentType)
	}

	// Decode the response body into the provided output parameter
	err = json.NewDecoder(res.Body).Decode(&output)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func PostDocumentUser(fileHeader *multipart.FileHeader) (*models.ResultPredict, error) {
	apiML := "http://localhost:8080/predict"

	// Open the file associated with the file header
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a buffer to write the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a part for the file
	part, err := writer.CreateFormFile("document", fileHeader.Filename)
	if err != nil {
		return nil, err
	}

	// Copy the file data to the part
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	// Close the writer to finalize the multipart form
	writer.Close()

	// Create the HTTP request
	req, err := http.NewRequest("POST", apiML, body)
	if err != nil {
		return nil, err
	}

	// Set the Content-Type header to match the form data
	req.Header.Set("Content-Type", writer.FormDataContentType())
	var result models.ResultPredict
	// Send the HTTP request
	res, err := SendHttpRequest(req, &result)
	if err != nil {

		return nil, err
	}

	// Check the HTTP response status code
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Failed to get a successful response from the API")
	}

	return &result, nil
}
