package model

const (
	GrafanaServiceAccountName      = "grafana-serviceaccount"
	GrafanaServiceName             = "grafana-service"
	GrafanaConfigName              = "grafana-config"
	GrafanaConfigFileName          = "grafana.ini"
	GrafanaIngressName             = "grafana-ingress"
	GrafanaRouteName               = "grafana-route"
	GrafanaDeploymentName          = "grafana-deployment"
	GrafanaPluginsVolumeName       = "grafana-plugins"
	GrafanaInitContainerName       = "grafana-plugins-init"
	GrafanaLogsVolumeName          = "grafana-logs"
	GrafanaDataVolumeName          = "grafana-data"
	GrafanaHealthEndpoint          = "/api/health"
	GrafanaPodLabel                = "grafana"
	LastConfigAnnotation           = "last-config"
	LastConfigEnvVar               = "LAST_CONFIG"
	GrafanaAdminSecretName         = "admin-credentials"
	DefaultAdminUser               = "admin"
	GrafanaAdminUserEnvVar         = "GF_SECURITY_ADMIN_USER"
	GrafanaAdminPasswordEnvVar     = "GF_SECURITY_ADMIN_PASSWORD"
	GrafanaHttpPort            int = 3000
)