package endpoints

//go:generate go run -tags=dev static_generate.go

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/kennep/timelapse/api"
	"github.com/kennep/timelapse/domain"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
)

// This is returned in error responses
type errorResponse struct {
	Message string `json:"message"`
}

func configureHandlerChain(r chi.Router) {
	r.Use(ApplicationContext())
	r.Use(chimiddleware.RealIP)
	r.Use(Logging())
	//r.Use(Authentication(users))
	r.Use(chimiddleware.Recoverer)

	r.Use(chimiddleware.Timeout(60 * time.Second))
}

type apiServer struct {
	users       *domain.Users
	staticFiles http.Handler
}

func Serve(users *domain.Users) error {
	listenAddr := os.Getenv("TIMELAPSE_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	server := apiServer{users, nil}

	r := chi.NewRouter()
	configureHandlerChain(r)

	ar := r.With(Authentication(users))

	ar.Get("/self", server.getCurrentUser)

	ar.Post("/projects", server.addProject)
	ar.Get("/projects", server.listProjects)
	ar.Get("/projects/{projectName}", server.getProject)
	ar.Put("/projects/{projectName}", server.updateProject)

	ar.Get("/entries", server.getUserTimeEntries)
	ar.Get("/projects/{projectName}/entries", server.getProjectTimeEntries)
	ar.Post("/projects/{projectName}/entries", server.addProjectTimeEntry)
	ar.Get("/projects/{projectName}/entries/{entryID}", server.getProjectTimeEntry)
	ar.Put("/projects/{projectName}/entries/{entryID}", server.updateProjectTimeEntry)

	r.Get("/*", server.serveStatic)

	server.staticFiles = http.FileServer(Assets)

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
	fields := RequestFields(r)
	fields["error"] = err
	log.WithFields(fields).Error(message)
	emitErrorResponse(rw, 500, "Internal Server Error")
}

func badRequest(rw http.ResponseWriter, r *http.Request, err error, message string, status int) {
	fields := RequestFields(r)
	if err != nil {
		fields["error"] = err
	}
	log.WithFields(fields).Warn(message)
	emitErrorResponse(rw, status, http.StatusText(status))
}

func validationError(rw http.ResponseWriter, r *http.Request, message string) {
	fields := RequestFields(r)
	log.WithFields(fields).Warnf("Validation error: %s", message)
	emitErrorResponse(rw, 400, message)
}

func notFoundError(rw http.ResponseWriter, r *http.Request, message string) {
	fields := RequestFields(r)
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
	user, err := s.users.GetOrCreateUserFromContext(ApplicationContextFromRequest(r))
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

	jsonResponse(rw, r, 200, mapUserToApi(user))
}

func (s *apiServer) addProject(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	var apiProject api.Project
	err := jsonRequest(rw, r, &apiProject)
	if err != nil {
		return
	}
	project := mapApiToProject(&apiProject)

	if project.Name == "" {
		validationError(rw, r, "required attribute: name")
		return
	}

	projectResult, err := user.GetProject(project.Name)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", project.Name))
		return
	}
	if projectResult != nil {
		validationError(rw, r, fmt.Sprintf("A project with this name already exists: %s", project.Name))
		return
	}

	projectResult, err = user.AddProject(project)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not create project %s", project.Name))
		return
	}
	jsonResponse(rw, r, 201, mapProjectToApi(projectResult))
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
	projectName, err := url.QueryUnescape(projectName)
	if err != nil {
		validationError(rw, r, err.Error())
		return
	}

	projectResult, err := user.GetProject(projectName)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", projectName))
		return
	}
	if projectResult == nil {
		notFoundError(rw, r, fmt.Sprintf("Project not found: %s", projectName))
		return
	}

	jsonResponse(rw, r, 200, mapProjectToApi(projectResult))
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
	projectName, err := url.QueryUnescape(projectName)
	if err != nil {
		validationError(rw, r, err.Error())
		return
	}
	var apiProject api.Project
	err = jsonRequest(rw, r, &apiProject)
	if err != nil {
		return
	}
	if apiProject.Name == "" {
		validationError(rw, r, "required attribute: name")
		return
	}

	projectResult, err := user.GetProject(projectName)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", projectName))
		return
	}
	if projectResult == nil {
		notFoundError(rw, r, fmt.Sprintf("Project not found: %s", projectName))
		return
	}

	mapApiToProjectDest(&apiProject, projectResult)

	err = projectResult.Save()
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not update project %s", projectName))
		return
	}
	jsonResponse(rw, r, 200, mapProjectToApi(projectResult))
}

func (s *apiServer) listProjects(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	projects, err := user.GetProjects()
	if err != nil {
		internalError(rw, r, err, "Could not list projects")
		return
	}

	jsonResponse(rw, r, 200, mapProjectsToApi(projects))
}

func (s *apiServer) getProjectTimeEntries(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	projectName := chi.URLParam(r, "projectName")
	if projectName == "" {
		validationError(rw, r, "project name in URL cannot be blank")
		return
	}
	projectName, err := url.QueryUnescape(projectName)
	if err != nil {
		validationError(rw, r, err.Error())
		return
	}

	project, err := user.GetProject(projectName)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", projectName))
		return
	}
	if project == nil {
		notFoundError(rw, r, fmt.Sprintf("Project not found: %s", projectName))
		return
	}

	entries, err := project.GetEntries()
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get time entries for project %s", projectName))
	}

	jsonResponse(rw, r, 200, mapTimeEntriesToApi(entries))
}

func (s *apiServer) getUserTimeEntries(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	entries, err := user.GetEntries()
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get time entries"))
	}

	jsonResponse(rw, r, 200, mapTimeEntriesToApi(entries))
}

func (s *apiServer) addProjectTimeEntry(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	var apiTimeEntry api.TimeEntry
	err := jsonRequest(rw, r, &apiTimeEntry)
	if err != nil {
		return
	}
	timeEntry := mapApiToTimeEntry(&apiTimeEntry)

	projectName := chi.URLParam(r, "projectName")
	if projectName == "" {
		validationError(rw, r, "project name in URL cannot be blank")
		return
	}
	projectName, err = url.QueryUnescape(projectName)
	if err != nil {
		validationError(rw, r, err.Error())
		return
	}

	project, err := user.GetProject(projectName)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", projectName))
		return
	}
	if project == nil {
		notFoundError(rw, r, fmt.Sprintf("Project not found: %s", projectName))
		return
	}

	newEntry, err := project.AddEntry(timeEntry)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Error while adding time entry"))
		return
	}

	jsonResponse(rw, r, 200, mapTimeEntryToApi(newEntry))
}

func (s *apiServer) updateProjectTimeEntry(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	var apiTimeEntry api.TimeEntry
	err := jsonRequest(rw, r, &apiTimeEntry)
	if err != nil {
		return
	}
	timeEntry := mapApiToTimeEntry(&apiTimeEntry)

	projectName := chi.URLParam(r, "projectName")
	if projectName == "" {
		validationError(rw, r, "project name in URL cannot be blank")
		return
	}
	projectName, err = url.QueryUnescape(projectName)
	if err != nil {
		validationError(rw, r, err.Error())
		return
	}

	project, err := user.GetProject(projectName)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", projectName))
		return
	}
	if project == nil {
		notFoundError(rw, r, fmt.Sprintf("Project not found: %s", projectName))
		return
	}

	entryID := chi.URLParam(r, "entryID")
	if entryID == "" {
		validationError(rw, r, "entry ID in URL cannot be blank")
		return
	}
	entryID, err = url.QueryUnescape(entryID)
	if err != nil {
		validationError(rw, r, err.Error())
		return
	}

	if timeEntry.ID == "" {
		timeEntry.ID = entryID
	}

	if entryID != timeEntry.ID {
		validationError(rw, r, "entry ID in URL does not match entry ID in body")
		return
	}

	updatedEntry, err := project.UpdateEntry(timeEntry)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Error while updating time entry"))
		return
	}

	jsonResponse(rw, r, 200, mapTimeEntryToApi(updatedEntry))
}

func (s *apiServer) getProjectTimeEntry(rw http.ResponseWriter, r *http.Request) {
	user := s.getUser(rw, r)
	if user == nil {
		return
	}

	projectName := chi.URLParam(r, "projectName")
	if projectName == "" {
		validationError(rw, r, "project name in URL cannot be blank")
		return
	}
	projectName, err := url.QueryUnescape(projectName)
	if err != nil {
		validationError(rw, r, err.Error())
		return
	}

	project, err := user.GetProject(projectName)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Could not get project by name %s", projectName))
		return
	}
	if project == nil {
		notFoundError(rw, r, fmt.Sprintf("Project not found: %s", projectName))
		return
	}

	entryID := chi.URLParam(r, "entryID")
	if entryID == "" {
		validationError(rw, r, "entry ID in URL cannot be blank")
		return
	}
	entryID, err = url.QueryUnescape(entryID)
	if err != nil {
		validationError(rw, r, err.Error())
		return
	}

	timeEntry, err := project.GetEntry(entryID)
	if err != nil {
		internalError(rw, r, err, fmt.Sprintf("Error while getting time entry"))
		return
	}
	if timeEntry == nil {
		notFoundError(rw, r, fmt.Sprintf("Time entry not found: %s/%s", projectName, entryID))
	}

	jsonResponse(rw, r, 200, mapTimeEntryToApi(timeEntry))
}

func (s *apiServer) serveStatic(rw http.ResponseWriter, r *http.Request) {
	s.staticFiles.ServeHTTP(rw, r)
}
