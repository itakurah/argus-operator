{{/*
Expand the name of the chart.
*/}}
{{- define "argus-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Kubernetes resource names 
*/}}
{{- define "argus-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Chart label.
*/}}
{{- define "argus-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "argus-operator.labels" -}}
helm.sh/chart: {{ include "argus-operator.chart" . }}
{{ include "argus-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "argus-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "argus-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Container image: <registry>/<repository>:<tag> when registry is set, else <repository>:<tag>
*/}}
{{- define "argus-operator.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- $repo := .Values.image.repository -}}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" (.Values.image.registry | trimSuffix "/") $repo $tag }}
{{- else }}
{{- printf "%s:%s" $repo $tag }}
{{- end }}
{{- end }}
