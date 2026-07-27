{{/*
Common labels applied to every resource in this chart.
*/}}
{{- define "serving-vllm.labels" -}}
app.kubernetes.io/name: serving-vllm
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Label selector shared by the Deployment and Service — also the selector
`miramar validate` polls on independently of the Helm release.
*/}}
{{- define "serving-vllm.selectorLabels" -}}
app.kubernetes.io/name: serving-vllm
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
