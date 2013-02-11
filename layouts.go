package main

import (
  "html/template"
  "io"
)

var emptyInterface interface{}

var headTemplate = template.Must(template.New("head").Parse(headTemplateHTML))
var headTemplateHTML = `
<html>
<head>
  <title>{{.title}}</title>
</head>
<body>
`

var tailTemplate = template.Must(template.New("tail").Parse(tailTemplateHTML))
var tailTemplateHTML = `
{{if .}}
<footer>{{.footer}}</footer>
{{end}}
</body>
</html>
`

var formFileUploadTemplate = template.Must(template.New("formFileUpload").Parse(formFileUploadTemplateHTML))
var formFileUploadTemplateHTML = `
<form enctype="multipart/form-data" action="" method="POST">                                            
  <input type="text" name="keywords" placeholder="keywords"><br/>
  <input type="file" name="filename" placeholder="filename"><br/>
  <input type="submit" value="Upload File"><br/>
</form>                                                                        
`

var listTemplate = template.Must(template.New("list").Parse(listTemplateHTML))
var listTemplateHTML = `
{{if .}}
<ul>
{{range .}}
  <li><a href="/f/{{.Filename}}">{{.Filename}}</a> - {{.Md5}}</li>
{{end}}
{{end}}
</ul>
`

func UploadPage(w io.Writer) (err error) {
  err = headTemplate.Execute(w, map[string]string{"title" : "FileSrv :: Upload"})
  if (err != nil) {
    return err
  }
  err = formFileUploadTemplate.Execute(w, &emptyInterface)
  if (err != nil) {
    return err
  }
  err = tailTemplate.Execute(w, map[string]string{})
  if (err != nil) {
    return err
  }
  return
}

func ListFilesPage(w io.Writer, files []File) (err error) {
  err = headTemplate.Execute(w, map[string]string{"title" : "FileSrv"})
  if (err != nil) {
    return err
  }
  err = listTemplate.Execute(w, files)
  if (err != nil) {
    return err
  }
  //err = tailTemplate.Execute(w, map[string]string{"footer" : "herp til you derp 2013"})
  err = tailTemplate.Execute(w, &emptyInterface)
  if (err != nil) {
    return err
  }
  return
}

