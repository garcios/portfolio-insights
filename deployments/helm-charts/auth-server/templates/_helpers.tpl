{{/*
Expand the name of the chart.
*/}}
{{- define "auth-server.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "auth-server.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "auth-server.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "auth-server.labels" -}}
helm.sh/chart: {{ include "auth-server.chart" . }}
{{ include "auth-server.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "auth-server.selectorLabels" -}}
app.kubernetes.io/name: {{ include "auth-server.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "auth-server.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "auth-server.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
PostgreSQL fullname
*/}}
{{- define "auth-server.postgres.fullname" -}}
{{- printf "%s-postgres" (include "auth-server.fullname" .) }}
{{- end }}

{{/*
PostgreSQL labels
*/}}
{{- define "auth-server.postgres.labels" -}}
{{ include "auth-server.labels" . }}
app.kubernetes.io/component: postgres
{{- end }}

{{/*
PostgreSQL selector labels
*/}}
{{- define "auth-server.postgres.selectorLabels" -}}
{{ include "auth-server.selectorLabels" . }}
app.kubernetes.io/component: postgres
{{- end }}

{{/*
Hydra Admin fullname
*/}}
{{- define "auth-server.hydraAdmin.fullname" -}}
{{- printf "%s-hydra-admin" (include "auth-server.fullname" .) }}
{{- end }}

{{/*
Hydra Admin labels
*/}}
{{- define "auth-server.hydraAdmin.labels" -}}
{{ include "auth-server.labels" . }}
app.kubernetes.io/component: hydra-admin
{{- end }}

{{/*
Hydra Admin selector labels
*/}}
{{- define "auth-server.hydraAdmin.selectorLabels" -}}
{{ include "auth-server.selectorLabels" . }}
app.kubernetes.io/component: hydra-admin
{{- end }}

{{/*
Hydra Public fullname
*/}}
{{- define "auth-server.hydraPublic.fullname" -}}
{{- printf "%s-hydra-public" (include "auth-server.fullname" .) }}
{{- end }}

{{/*
Hydra Public labels
*/}}
{{- define "auth-server.hydraPublic.labels" -}}
{{ include "auth-server.labels" . }}
app.kubernetes.io/component: hydra-public
{{- end }}

{{/*
Hydra Public selector labels
*/}}
{{- define "auth-server.hydraPublic.selectorLabels" -}}
{{ include "auth-server.selectorLabels" . }}
app.kubernetes.io/component: hydra-public
{{- end }}

{{/*
Login Consent Provider fullname
*/}}
{{- define "auth-server.loginConsentProvider.fullname" -}}
{{- printf "%s-login-consent-provider" (include "auth-server.fullname" .) }}
{{- end }}

{{/*
Login Consent Provider labels
*/}}
{{- define "auth-server.loginConsentProvider.labels" -}}
{{ include "auth-server.labels" . }}
app.kubernetes.io/component: login-consent-provider
{{- end }}

{{/*
Login Consent Provider selector labels
*/}}
{{- define "auth-server.loginConsentProvider.selectorLabels" -}}
{{ include "auth-server.selectorLabels" . }}
app.kubernetes.io/component: login-consent-provider
{{- end }}

{{/*
Hydra Migrate Job fullname
*/}}
{{- define "auth-server.hydraMigrate.fullname" -}}
{{- printf "%s-hydra-migrate" (include "auth-server.fullname" .) }}
{{- end }}

{{/*
Hydra Migrate Job labels
*/}}
{{- define "auth-server.hydraMigrate.labels" -}}
{{ include "auth-server.labels" . }}
app.kubernetes.io/component: hydra-migrate
{{- end }}

{{/*
Database DSN
*/}}
{{- define "auth-server.dsn" -}}
{{- printf "postgres://%s:%s@%s:%d/%s?sslmode=disable" .Values.postgres.auth.username .Values.postgres.auth.password (include "auth-server.postgres.fullname" .) (int .Values.postgres.service.port) .Values.postgres.auth.database }}
{{- end }}
