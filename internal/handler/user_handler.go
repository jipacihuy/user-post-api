package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"user-post-api/internal/model"
	"user-post-api/internal/service"
	"user-post-api/internal/utils"
	"user-post-api/pkg/jwt"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Service *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{Service: svc}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.RespondError(w, http.StatusBadRequest, errors[0].Translate(nil))
		return
	}

	id, err := h.Service.CreateUser(user)
	if err != nil {
		utils.RespondError(w, http.StatusConflict, err.Error())
		return
	}
	user.ID = id
	user.Password = ""
	utils.RespondSuccess(w, http.StatusCreated, user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.RespondError(w, http.StatusBadRequest, errors[0].Translate(nil))
		return
	}

	user, err := h.Service.GetUserByEmail(req.Email)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, _ := jwt.GenerateToken(user.ID)
	user.Password = ""
	utils.RespondSuccess(w, http.StatusOK, model.LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	user, err := h.Service.GetUserByID(userID)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "user not found")
		return
	}
	user.Password = ""
	utils.RespondSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Service.GetAllUsers()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondSuccess(w, http.StatusOK, users)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid ID")
		return
	}
	user, err := h.Service.GetUserByID(id)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "user not found")
		return
	}
	user.Password = ""
	utils.RespondSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.RespondError(w, http.StatusBadRequest, errors[0].Translate(nil))
		return
	}

	id, err := h.Service.CreateUser(user)
	if err != nil {
		utils.RespondError(w, http.StatusConflict, err.Error())
		return
	}
	user.ID = id
	user.Password = ""
	utils.RespondSuccess(w, http.StatusCreated, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := h.Service.UpdateUser(id, user); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	user.ID = id
	user.Password = ""
	utils.RespondSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	if err := h.Service.DeleteUser(id); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	utils.RespondSuccess(w, http.StatusOK, map[string]string{"message": "user deleted"})
}
