package controllers

import (
	"net/http"
	"projectfiber/models"
	"projectfiber/repository"

	//"projectfiber/usecase"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// fungsi ini bertujuan untuk mengambil seluruh isi dalam database dan dimunculkan
func GetTodoLists(c *fiber.Ctx) error {

	var todolists []models.TodoList //variabel local yang akan mengambil isi database

	repository.DB.Find(&todolists)
	return c.JSON(todolists)
}

func GetTodoList(c *fiber.Ctx) error {

	// get post id dari parameter
	id := c.Params("id") //jika parameter pada http sama dengan parameter pada database
	var todolist models.TodoList

	if err := repository.DB.First(&todolist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "Data tidak ditemukan", //jika terdapat perbedaan antara parameter yang dimasukkan dengan yang ada didatabase
			}) //maka akan melempar status not found
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Unexpected Error", //jika terdapat error namun bukan pada parameter alias di db, maka akan error
		})
	}
	return c.JSON(todolist) //jika parameter sesuai, maka akan menampilkan isi db yang sesuai dengan parameter inputtan
}

// memasukkan input user kedalam database
func CreateTodoList(c *fiber.Ctx) error {

	var todolist models.TodoList

	if err := c.BodyParser(&todolist); err != nil {
		return err //jika saat mengekstrak body terdapat error maka akan melempar error
	}
	if err := repository.DB.Create(&todolist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(), //jika terdapat error dalam database fungsi create maka akan melempar internal server error
		})
	}
	return c.JSON(todolist) //jika tidak maka akan melempar berhasil dengan melempar isi database
}

func UpdateTodoList(c *fiber.Ctx) error {

	//pencocokan parameter
	id := c.Params("id")
	// cari post memakai id
	var todolist models.TodoList

	if err := c.BodyParser(&todolist); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(), //jika terdapat error pada db karena perbedaan parameter maka akan melempar pesan error
		})
	}
	err := repository.DB.Model(models.TodoList{}).Where("id", id).Debug().Updates(&todolist).Error
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Tidak bisa memperbarui data: " + err.Error(), //jika terdapat error karena ketidaksesuaian parameter maka akan melempar pesan
		})
	}
	return c.JSON(fiber.Map{
		"message": "Data berhasil diperbarui",
	}) //jika berhasil
}

func DeleteTodoList(c *fiber.Ctx) error {
	// get post id dari parameter
	id := c.Params("id")
	var todolist models.TodoList

	if repository.DB.Delete(&todolist, id).RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "Data sudah dihapus sebelumnya.", //jika terdapat db yang sudah dihapus maka akan melempar pesan
		})
	}
	return c.JSON(fiber.Map{
		"message": "Data berhasil dihapus",
	})
}
