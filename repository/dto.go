package repository

import (
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type (
	identity struct {
		Issuer    string `bson:"issuer"`
		SubjectID string `bson:"subjectid"`
		Email     string `bson:"email"`
	}

	user struct {
		ID         primitive.ObjectID `bson:"_id"`
		Identities []identity         `bson:"identities"`
	}

	project struct {
		ID          primitive.ObjectID `bson:"_id"`
		UserID      primitive.ObjectID `bson:"userid"`
		Name        string             `bson:"name"`
		Description string             `bson:"description"`
		Billable    bool               `bson:"billable"`
	}

	timeEntry struct {
		ID        primitive.ObjectID `bson:"_id"`
		ProjectID primitive.ObjectID `bson:"projectid"`
		UserID    primitive.ObjectID `bson:"userid"`
		Start     time.Time          `bson:"start"`
		End       time.Time          `bson:"end"`
		Breaks    time.Duration      `bson:"breaks"`
		Comment   string             `bson:"comment"`
	}

	dayEntry struct {
		ID     primitive.ObjectID `bson:"_id"`
		Type   string             `bson:"type"`
		UserID string             `bson:"userid"`
		Date   time.Time          `bson:"date"`
	}
)
