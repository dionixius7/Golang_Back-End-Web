package models

import (
	"time"

	"github.com/google/uuid"
)

type TodoList struct {
	ID          int     `gorm:"primaryKey" json:"id"`
	Title       *string `gorm:"varchar(300)" json:"title"`
	Description *string `gorm:"text" json:"description"`
	Progress    *bool   `gorm:"default:false" json:"progress"`
	//Favorite    *bool   `gorm:"default:false" json:"favorite"`
}

type Users struct {
	ID   uuid.UUID `gorm:"type:char(36);default:uuid()" json:"id"`
	Name string    `gorm:"type:varchar(300);" json:"name"`
	//Username     string     `gorm:"type:varchar(300);unique;" json:"username"`
	Email        string     `gorm:"type:varchar(300);unique;" json:"email"`
	Password     string     `gorm:"type:text;" json:"password"`
	AccStatus    bool       `gorm:"default:true;" json:"acc_status"`
	RefreshToken string     `gorm:"type:text;" json:"refresh_token"`
	AccessToken  string     `gorm:"type:text;" json:"access_token"`
	CreatedAt    time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type JobUser struct {
	ID      uuid.UUID `gorm:"type:char(36);default:uuid()" json:"id"`
	EmailID uuid.UUID `json:"email_id"`
	Job     string    `gorm:"type:varchar(36);" json:"job"`
}

type ResultPredict struct {
	HasilPrediksi string `json:"hasil_prediksi,omitempty"`
}

type ReqSignUp struct {
	Name  string `gorm:"type:varchar(300);" json:"name"`
	Email string `gorm:"type:varchar(300);unique;" json:"email"`
	//Username string `json:"username"`
	Password string `gorm:"type:text;" json:"password"`
}

type LoginPayLoad struct {
	//Username string `gorm:"type:varchar;unique" json:"username" validate:"required"`
	//TODO email
	Email    string `gorm:"type:varchar(300);unique;" json:"email"`
	Password string `gorm:"type:text;" json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRes struct {
	AccessToken string `json:"access_token"`
}

type SessionUser struct {
	ID           uuid.UUID `json:"id"`
	EmailID      uuid.UUID `json:"email_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    *bool     `gorm:"default:false;" json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `gorm:"default:now();" json:"created_at"`
}
