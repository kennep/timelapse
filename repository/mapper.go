package repository

import (
	"encoding/base64"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"github.com/kennep/timelapse/domain"
)

func idToString(id primitive.ObjectID) string {
	return base64.RawStdEncoding.EncodeToString(id[:])
}

func stringToID(id string) (primitive.ObjectID, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(id)
	if err != nil {
		return primitive.NilObjectID, err
	}

	var objID primitive.ObjectID
	copy(objID[:], decoded)

	return objID, nil
}

func stringsToIDs(ids ...string) ([]primitive.ObjectID, error) {
	var result []primitive.ObjectID
	for _, id := range ids {
		objID, err := stringToID(id)
		if err != nil {
			return nil, err
		}
		result = append(result, objID)
	}
	return result, nil
}

func mapUserToDomain(in *user) *domain.User {
	out := domain.User{
		ID: idToString(in.ID),
	}
	for _, identity := range in.Identities {
		out.Identities = append(out.Identities,
			domain.Identity{
				Issuer:    identity.Issuer,
				SubjectID: identity.SubjectID,
				Email:     identity.Email,
			})
	}
	return &out
}

func mapProjectToDomain(in *project) *domain.Project {
	return &domain.Project{
		ID:          idToString(in.ID),
		UserID:      idToString(in.UserID),
		Name:        in.Name,
		Description: in.Description,
		Billable:    in.Billable,
	}
}

func mapProjectFromDomain(in *domain.Project) (*project, error) {
	ids, err := stringsToIDs(in.ID, in.UserID)
	if err != nil {
		return nil, err
	}
	return &project{
		ID:          ids[0],
		UserID:      ids[1],
		Name:        in.Name,
		Description: in.Description,
		Billable:    in.Billable,
	}, nil
}

func mapTimeEntryToDomain(in *timeEntry) *domain.TimeEntry {
	return &domain.TimeEntry{
		ID:        idToString(in.ID),
		ProjectID: idToString(in.ProjectID),
		UserID:    idToString(in.UserID),
		Start:     in.Start,
		End:       in.End,
		Breaks:    in.Breaks,
		Comment:   in.Comment,
	}
}

func mapTimeEntryFromDomain(in *domain.TimeEntry) (*timeEntry, error) {
	ids, err := stringsToIDs(in.ID, in.ProjectID, in.UserID)
	if err != nil {
		return nil, err
	}
	return &timeEntry{
		ID:        ids[0],
		ProjectID: ids[1],
		UserID:    ids[2],
		Start:     in.Start,
		End:       in.End,
		Breaks:    in.Breaks,
		Comment:   in.Comment,
	}, nil
}
