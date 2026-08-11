package artifact

import "github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"

func modelsAgent(codename string) models.Agent { return models.Agent{Codename: codename, Tier: 5} }
