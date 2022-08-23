package notification

import (
	"github.com/cresta/atlantis-drift-detection/internal/testhelper"
	"net/http"
	"testing"
)

func TestSlackWebhook_ExtraWorkspaceInRemote(t *testing.T) {
	testhelper.ReadEnvFile(t, "../../")
	wh := NewSlackWebhook(testhelper.EnvOrSkip(t, "SLACK_WEBHOOK_URL"), http.DefaultClient)
	genericNotificationTest(t, wh)
}
