package endpoints

import (
	"github.com/kennep/timelapse/api"
	"github.com/kennep/timelapse/domain"
)

func mapUserToApi(user *domain.User) *api.User {
	var result api.User
	result.ID = user.ID
	for _, identity := range user.Identities {
		var apiIdentity api.Identity
		apiIdentity.Issuer = identity.Issuer
		apiIdentity.SubjectID = identity.SubjectID
		apiIdentity.Email = identity.Email
		result.Identities = append(result.Identities, apiIdentity)
	}
	return &result
}

func mapProjectToApi(project *domain.Project) *api.Project {
	var result api.Project

	result.Name = project.Name
	result.Description = project.Description
	result.Billable = project.Billable

	return &result
}

func mapApiToProject(project *api.Project) *domain.Project {
	var result domain.Project

	mapApiToProjectDest(project, &result)

	return &result
}

func mapApiToProjectDest(project *api.Project, result *domain.Project) {
	result.Name = project.Name
	result.Description = project.Description
	result.Billable = project.Billable
}

func mapProjectsToApi(projects []*domain.Project) []*api.Project {
	var result []*api.Project
	for _, project := range projects {
		apiProject := mapProjectToApi(project)
		result = append(result, apiProject)
	}
	return result
}

func mapTimeEntryToApi(entry *domain.TimeEntry) *api.TimeEntry {
	var result api.TimeEntry

	result.ProjectName = entry.Project.Name
	result.Type = entry.Type
	result.Start = entry.Start
	result.End = entry.End
	result.Breaks = entry.Breaks
	result.Comment = entry.Comment

	return &result
}

func mapApiToTimeEntry(entry *api.TimeEntry) *domain.TimeEntry {
	var result domain.TimeEntry

	result.Type = entry.Type
	result.Start = entry.Start
	result.End = entry.End
	result.Breaks = entry.Breaks
	result.Comment = entry.Comment

	return &result
}

func mapTimeEntriesToApi(entries []*domain.TimeEntry) []*api.TimeEntry {
	var result []*api.TimeEntry
	for _, entry := range entries {
		apiEntry := mapTimeEntryToApi(entry)
		result = append(result, apiEntry)
	}
	return result
}
