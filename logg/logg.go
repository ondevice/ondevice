package logg

import (
	"github.com/ondevice/ondevice/util"
	"github.com/sirupsen/logrus"
)

// FailWithAPIError -- call Fatal with a nice error message matching the API error we got
func FailWithAPIError(err util.APIError) {
	if err.Code() == util.NotFoundError {
		logrus.Fatal("Not found: ", err.Error())
	} else if err.Code() == util.AuthenticationError {
		logrus.Fatal("Authentication failed (try running ondevice login)")
	} else if err.Code() == util.ForbiddenError {
		logrus.Fatal("Access denied (are you sure your API key has the required roles?)")
	} else {
		logrus.Fatalf("Unexpected API error (code %d): %s", err.Code(), err.Error())
	}
}
