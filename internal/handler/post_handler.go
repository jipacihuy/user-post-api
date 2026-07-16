package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"user-post-api/internal/model"
	"user-post-api/internal/service"
	"user-post-api/internal/utils"
	"github.com/gorilla/mux"
)

type PostHandler struct {
	Service *service.PostService
}

func NewPostHandler(svc *service.PostService) *PostHandler {
	return &PostHandler{Service: svc}
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.Service.GetAllPosts()
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondSuccess(w, http.StatusOK, posts)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid ID")
		return
	}
	post, err := h.Service.GetPostByID(id)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "post not found")
		return
	}
	utils.RespondSuccess(w, http.StatusOK, post)
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var post model.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid request")
		return
	}
	id, err := h.Service.CreatePost(post)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	post.ID = id
	utils.RespondSuccess(w, http.StatusCreated, post)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid ID")
		return
	}
	if err := h.Service.DeletePost(id); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondSuccess(w, http.StatusOK, map[string]string{"message": "post deleted"})
}