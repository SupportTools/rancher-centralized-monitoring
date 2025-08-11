{{/*
Expand the name of the chart.
*/}}
{{- define "rancher-centralized-monitoring.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "rancher-centralized-monitoring.fullname" -}}
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
{{- define "rancher-centralized-monitoring.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "rancher-centralized-monitoring.labels" -}}
helm.sh/chart: {{ include "rancher-centralized-monitoring.chart" . }}
{{ include "rancher-centralized-monitoring.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "rancher-centralized-monitoring.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rancher-centralized-monitoring.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "rancher-centralized-monitoring.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "rancher-centralized-monitoring.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create secret name for Rancher API credentials
*/}}
{{- define "rancher-centralized-monitoring.secretName" -}}
{{- if .Values.rancher.auth.existingSecret }}
{{- .Values.rancher.auth.existingSecret }}
{{- else }}
{{- include "rancher-centralized-monitoring.fullname" . }}-credentials
{{- end }}
{{- end }}