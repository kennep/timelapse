package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kennep/timelapse/domain"
	"github.com/kennep/timelapse/middleware"
	"github.com/kennep/timelapse/repository"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
)

// This is returned in error responses
type errorResponse struct {
	Message string `json:"message"`
}

func configureHandlerChain(repo *repository.TimelapseRepository, r chi.Router) {
	r.Use(middleware.ApplicationContext())
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logging())
	r.Use(middleware.Authentication(repo))
	r.Use(chimiddleware.Recoverer)

	r.Use(chimiddleware.Timeout(60 * time.Second))
}

type apiServer struct {
	repo *repository.TimelapseRepository
}

func Serve(repo *repository.TimelapseRepository) error {
	listenAddr := os.Getenv("TIMELAPSE_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	server := apiServer{repo}

	r := chi.NewRouter()
	configureHandlerChain(repo, r)

	r.Get("/self", server.getCurrentUser)

	r.Post("/projects", server.addProject)
	r.Get("/projects/{projectName}", server.getProject)
	r.Put("/projects/{projectName}", server.updateProject)

	log.WithFields(log.Fields{"address": listenAddr}).Info("Timelapse server listening")
	if err := http.ListenAndServe(listenAddr, r); err != nil {
		return err
	}

	return nil
}

func emitErrorResponse(rw http.ResponseWriter, statusCode int, errorMessage string) {
	errResponse := errorResponse{errorMessage}
	body, err := json.Marshal(errResponse)
	if err != nil {
		// oh dear... errors when marshaling the error response. We'll write the message as plain text instead
		rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		rw.WriteHeader(statusCode)
		rw.Write([]byte(errorMessage))
	} else {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(statusCode)
		rw.Write(body)
	}
}

func internalError(rw http.ResponseWriter, r *http.Request, err error, message string) {
	fields := middleware.RequestFields(r)
	fields["error"] = err
	log.WithFields(fields).Error(message)
	emitErrorResponse(rw, 500, "Internal Server Error")
}

func badRequest(rw http.ResponseWriter, r *http.Request, err error, message string, status int) {
	fields := middleware.RequestFields(r)
	if err != nil {
		fields["error"] = err
	}
	log.WithFields(fields).Warn(message)
	emitErrorResponse(rw, status, http.StatusText(status))
}

func validationError(rw http.ResponseWriter, r *http.Request, message string) {
	fields := middleware.RequestFields(r)
	log.WithFields(fields).Warnf("Validation error: %s", message)
	emitErrorResponse(rw, 400, message)
}

func notFoundError(rw http.ResponseWriter, r *http.Request, message string) {
	fields := middleware.RequestFields(r)
	log.WithFields(fields).Warnf("Not found error: %s", message)
	emitErrorResponse(rw, 404, message)
}

func jsonResponse(rw http.ResponseWriter, r *http.Request, statusCode int, doc interface{}) {
	body, err := json.Marshal(doc)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not serialize to json: %v", doc))
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	rw.Write(body)
}

func jsonRequest(rw http.ResponseWriter, r *http.Request, target interface{}) error {
	contentType := r.Header.Get("Content-Type")
	mimeType, mimeParams, _ := mime.ParseMediaType(contentType)
	if mimeType != "" && mimeType != "application/json" {
		badRequest(rw, r, nil, fmt.Sprintf("Invalid content type: %s", mimeType), 415)
		return errors.New("Invalid content type")
	}
	if charset, ok := mimeParams["charset"]; ok && strings.ToLower(charset) != "utf-8" {
		badRequest(rw, r, nil, fmt.Sprintf("Invalid charset: %s", charset), 415)
		return errors.New("Invalid charset")
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		badRequest(rw, r, err, "Could not read request body", 400)
		return err
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		badRequest(rw, r, err, "Could not unmarshal body", 400)
		return err
	}

	return nil
}

func (s *apiServer) getUser(rw http.ResponseWriter, r *http.Request) *domain.User {
	user, err := s.repo.CreateUserFromContext(middleware.ApplicationContextFromRequest(r))
	if err != nil {
		internalError(rw, r, err, "Could not construct user from application context")
		return nil
	}
	return user
}

func (s *apiServer) getCurrentUser(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	jsonResponse(rw, r, 200, user)
}

func (s *apiServer) addProject(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	var project domain.Project
	err := jsonRequest(rw, r, &project)
	if err != nil {
		return
	}

	if project.Name == "" {
		validationError(rw, r, "required attribute: name")
		return
	}

	projectResult, err := s.repo.GetProject(user, project.Name)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", project.Name))
		return
	}
	if projectResult != nil {
		validationError(rw, r, fmt.Sprintf("A project with this name already exists: %s", project.Name))
		return
	}

	projectResult, err = s.repo.AddProject(user, &project)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not create project %s", project.Name))
		return
	}
	jsonResponse(rw, r, 201, projectResult)
}

func (s *apiServer) getProject(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	projectName := chi.URLParam(r, "projectName")
	if projectName == "" {
		validationError(rw, r, "project name in URL cannot be blank")
		return
	}

	projectResult, err := s.repo.GetProject(user, projectName)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", projectName))
		return
	}
	if projectResult == nil {
		notFoundError(rw, r, fmt.Sprintf("Project not found: %s", projectName))
		return
	}

	jsonResponse(rw, r, 200, projectResult)
}

func (s *apiServer) updateProject(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	projectName := chi.URLParam(r, "projectName")
	if projectName == "" {
		validationError(rw, r, "project name in URL cannot be blank")
		return
	}
	var project domain.Project
	err := jsonRequest(rw, r, &project)
	if err != nil {
		return
	}

	if project.Name == "" {
		validationError(rw, r, "required attribute: name")
		return
	}

	projectResult, err := s.repo.GetProject(user, projectName)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", projectName))
		return
	}
	if projectResult == nil {
		notFoundError(rw, r, fmt.Sprintf("Project not found: %s", projectName))
		return
	}

	projectResult, err = s.repo.UpdateProject(user, projectName, &project)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not update project %s", projectName))
		return
	}
	jsonResponse(rw, r, 200, projectResult)
}
