package main

import (
	"fmt"
	"io"
	"text/template"

	humanize "github.com/dustin/go-humanize"
	"github.com/vbatts/imgsrv/types"
)

var emptyInterface interface{}

var headTemplate = template.Must(template.New("head").Parse(headTemplateHTML))
var headTemplateHTML = `
<html>
<head>
  <link href="/assets/bootstrap.css" media="screen" rel="stylesheet" type="text/css" />

  <script src="/assets/jquery.js" type="text/javascript" ></script>
  <script src="/assets/jqud.js" type="text/javascript" ></script>
  <script src="/assets/bootstrap.js" type="text/javascript" ></script>

  <title>{{.title}}</title>
</head>
<body>
`

var navbarTemplate = template.Must(template.New("navbar").Parse(navbarTemplateHTML))
var navbarTemplateHTML = `
  <div class="navbar navbar-inverse navbar-fixed-top">
    <div class="navbar-inner">
      <div class="container-fluid">
        <a class="btn btn-navbar" data-toggle="collapse" data-target=".nav-collapse">
          <span class="icon-bar"></span>
          <span class="icon-bar"></span>
          <span class="icon-bar"></span>
        </a>
        <a class="brand" href="/">filesrv</a>
        <div class="nav-collapse collapse">
          <p class="navbar-text pull-right">
          </p>
          <ul class="nav">
            <li><a href="/">Home</a></li>
            <li><a href="/upload">Upload</a></li>
            <li><a href="/urlie">URLie</a></li>
            <li><a href="/all">All</a></li>
          </ul>
          <div class="dropdown nav pull-right">
            <a role="button" data-toggle="dropdown" href="#">Other<b class="caret"></b></a>
            <ul class="dropdown-menu" role="menu" aria-labelledby="dLabel">
              <li><a href="/k/">Keywords</a></li>
              <li><a href="/ext/">File ext</a></li>
              <li><a href="/md5/">MD5s</a></li>
            </ul>
          </div> <!-- dropdown -->
        </div> <!-- nav-collapse -->
      </div> <!-- container-fluid -->
    </div> <!-- navbar-inner -->
  </div> <!-- navbar top -->
  <script>$('.dropdown-toggle').dropdown()</script>

`
var containerBeginTemplate = template.Must(template.New("containerBegin").Parse(containerBeginTemplateHTML))
var containerBeginTemplateHTML = `
<div class="container-fluid">
  <hr>
  <hr>
  <div class="row-fluid">
`

var tailTemplate = template.Must(template.New("tail").Parse(tailTemplateHTML))
var tailTemplateHTML = `
  </div>
  <hr>
</div>
{{if .}}
<footer>{{.footer}}</footer>
{{end}}
</body>
</html>
`

var formDeleteFileTemplate = template.Must(template.New("formDeleteFile").Parse(formDeleteFileTemplateHTML))
var formDeleteFileTemplateHTML = `
<div class="span9">
<div class="hero-unit">
  <h3>Get file from URL</h3>
{{if .}}
<table>
<tr>
<b>Are you sure?</b>
</tr>
<br/>
<tr>
<td>
<a role="button" href="/v/{{.}}">no!</a>
<br/>
<a role="button" href="/f/{{.}}?delete=true&confirm=true">yes! delete!</a>
</td>
</tr>
</table>
{{else}}
<p>
<b>ERROR: No File provided!</b>
</p>
{{end}}
</div>{{/* hero-unit */}}
</div>{{/* span9 */}}
`

var formGetUrlTemplate = template.Must(template.New("formGetUrl").Parse(formGetUrlTemplateHTML))
var formGetUrlTemplateHTML = `
<div class="span9">
<div class="hero-unit">
  <h3>Get file from URL</h3>
<form enctype="multipart/form-data" action="/urlie" method="POST">
  <table>
    <tr>
  <td>
      <input type="text" name="url" placeholder="file URL"><br/>
      <input type="text" name="keywords" placeholder="keywords"><i>(comma seperatated, no spaces)</i><br/>
      <input type="checkbox" name="rand" value="true">Randomize filename
  </td>
    </tr>
    <tr>
    <td>
      <input type="submit" value="Fetch File"><br/>
  </td>
    </tr>
  </td>
  </table>
</form>
</div>{{/* hero-unit */}}
</div>{{/* span9 */}}
`

var formFileUploadTemplate = template.Must(template.New("formFileUpload").Parse(formFileUploadTemplateHTML))
var formFileUploadTemplateHTML = `
<div class="span9">
<div class="hero-unit">
  <h3>Upload a File</h3>
<form enctype="multipart/form-data" action="/upload" method="POST">
  <table>
    <tr>
  <td>
      <input type="file" name="filename" placeholder="filename"><br/>
      <input type="text" name="keywords" placeholder="keywords"><i>(comma seperatated, no spaces)</i><br/>
      <input type="checkbox" name="rand" value="true">Randomize filename
  </td>
    </tr>
    <tr>
    <td>
      <input type="submit" value="Upload File"><br/>
  </td>
    </tr>
  </td>
  </table>
</form>
</div>{{/* hero-unit */}}
</div>{{/* span9 */}}
`

var listTemplate = template.Must(template.New("list").Parse(listTemplateHTML))
var listTemplateHTML = `
{{if .}}
<ul>
{{range .}}
<li>
<a href="/v/{{.Filename}}">{{.Filename}}</a>
[keywords:{{range $key := .Metadata.Keywords}} <a href="/k/{{$key}}">{{$key}}</a>{{end}}]
[md5: <a href="/md5/{{.Md5}}">{{.Md5 | printf "%8.8s"}}...</a>]</li>
{{end}}
</ul>
{{end}}
`

var tagcloudTemplate = template.Must(template.New("tagcloud").Parse(tagcloudTemplateHTML))
var tagcloudTemplateHTML = `
{{if .}}
<div id="tagCloud">
{{range .}}
<a href="/{{.Root}}/{{.Id}}" rel="{{.Value}}">{{.Id}}</a>
{{end}}
</div>
{{end}}

<script>
$.fn.tagcloud.defaults = {
  size: {start: 9, end: 40, unit: 'pt'},
  color: {start: '#007ab7', end: '#e55b00'}
};

$(function () {
  $('#tagCloud a').tagcloud();
});
</script>
`

var fileViewImageTemplate = template.Must(template.New("file").Parse(fileViewImageTemplateHTML))
var fileViewImageTemplateHTML = `
{{if .}}
<a href="/f/{{.Filename}}"><img src="/f/{{.Filename}}"></a>
{{end}}
`

var fileViewAudioTemplate = template.Must(template.New("file").Parse(fileViewAudioTemplateHTML))
var fileViewAudioTemplateHTML = `
{{if .}}
<a href="/f/{{.Filename}}">
<audio controls>
  <source src="/f/{{.Filename}}" autoplay/>
  Your browser does not support the video tag.
</audio>
</a>
{{end}}
`

var fileViewVideoTemplate = template.Must(template.New("file").Parse(fileViewVideoTemplateHTML))
var fileViewVideoTemplateHTML = `
{{if .}}
<a href="/f/{{.Filename}}">
<video width="320" height="240" controls>
  <source src="/f/{{.Filename}}" autoplay/>
  Your browser does not support the video tag.
</video>
</a>
{{end}}
`

var fileViewTemplate = template.Must(template.New("file").Parse(fileViewTemplateHTML))
var fileViewTemplateHTML = `
{{if .}}
<a href="/f/{{.Filename}}">{{.Filename}}</a>
{{end}}
`

var funcs = template.FuncMap{
	"humanBytes": humanize.Bytes,
	"humanTime":  humanize.Time,
}

var fileViewInfoTemplate = template.Must(template.New("file").Funcs(funcs).Parse(fileViewInfoTemplateHTML))
var fileViewInfoTemplateHTML = `
{{if .}}
<br/>
[keywords:{{range $key := .Metadata.Keywords}} <a href="/k/{{$key}}">{{$key}}</a>{{end}}]
<br/>
[md5: <a href="/md5/{{.Md5}}">{{.Md5}}</a>]
<br/>
[size: {{humanBytes .Length}}]
<br/>
[UploadDate: {{.Metadata.TimeStamp}} ({{humanTime .Metadata.TimeStamp}})]
<br/>
[<a href="/f/{{.Filename}}?delete=true">Delete</a>]
{{end}}
`

func DeleteFilePage(w io.Writer, filename string) (err error) {
	err = headTemplate.Execute(w, map[string]string{"title": "FileSrv :: delete"})
	if err != nil {
		return err
	}
	err = navbarTemplate.Execute(w, nil)
	if err != nil {
		return err
	}
	err = containerBeginTemplate.Execute(w, nil)
	if err != nil {
		return err
	}

	err = formDeleteFileTemplate.Execute(w, &filename)
	if err != nil {
		return err
	}

	err = tailTemplate.Execute(w, map[string]string{"footer": fmt.Sprintf("Version: %s", VERSION)})
	if err != nil {
		return err
	}
	return
}

func UrliePage(w io.Writer) (err error) {
	err = headTemplate.Execute(w, map[string]string{"title": "FileSrv :: URLie"})
	if err != nil {
		return err
	}
	err = navbarTemplate.Execute(w, nil)
	if err != nil {
		return err
	}
	err = containerBeginTemplate.Execute(w, nil)
	if err != nil {
		return err
	}
	err = formGetUrlTemplate.Execute(w, &emptyInterface)
	if err != nil {
		return err
	}
	err = tailTemplate.Execute(w, map[string]string{"footer": fmt.Sprintf("Version: %s", VERSION)})
	if err != nil {
		return err
	}
	return
}

func UploadPage(w io.Writer) (err error) {
	err = headTemplate.Execute(w, map[string]string{"title": "FileSrv :: Upload"})
	if err != nil {
		return err
	}
	err = navbarTemplate.Execute(w, nil)
	if err != nil {
		return err
	}
	err = containerBeginTemplate.Execute(w, nil)
	if err != nil {
		return err
	}

	// main context of this page
	err = formFileUploadTemplate.Execute(w, &emptyInterface)
	if err != nil {
		return err
	}

	err = tailTemplate.Execute(w, map[string]string{"footer": fmt.Sprintf("Version: %s", VERSION)})
	if err != nil {
		return err
	}
	return
}

func ImageViewPage(w io.Writer, file types.File) (err error) {
	err = headTemplate.Execute(w, map[string]string{"title": "FileSrv"})
	if err != nil {
		return err
	}
	err = navbarTemplate.Execute(w, nil)
	if err != nil {
		return err
	}
	err = containerBeginTemplate.Execute(w, nil)
	if err != nil {
		return err
	}

	if file.IsImage() {
		err = fileViewImageTemplate.Execute(w, file)
	} else if file.IsAudio() {
		err = fileViewAudioTemplate.Execute(w, file)
	} else if file.IsVideo() {
		err = fileViewVideoTemplate.Execute(w, file)
	} else {
		err = fileViewTemplate.Execute(w, file)
	}
	if err != nil {
		return err
	}
	err = fileViewInfoTemplate.Execute(w, file)
	if err != nil {
		return err
	}

	err = tailTemplate.Execute(w, map[string]string{"footer": fmt.Sprintf("Version: %s", VERSION)})
	if err != nil {
		return err
	}
	return
}

func ListFilesPage(w io.Writer, files []types.File) (err error) {
	err = headTemplate.Execute(w, map[string]string{"title": "FileSrv"})
	if err != nil {
		return err
	}
	err = navbarTemplate.Execute(w, nil)
	if err != nil {
		return err
	}
	err = containerBeginTemplate.Execute(w, nil)
	if err != nil {
		return err
	}

	// main context of this page
	err = listTemplate.Execute(w, files)
	if err != nil {
		return err
	}

	err = tailTemplate.Execute(w, map[string]string{"footer": fmt.Sprintf("Version: %s", VERSION)})
	if err != nil {
		return err
	}
	return
}

func ListTagCloudPage(w io.Writer, ic []types.IdCount) (err error) {
	err = headTemplate.Execute(w, map[string]string{"title": "FileSrv"})
	if err != nil {
		return err
	}
	err = navbarTemplate.Execute(w, nil)
	if err != nil {
		return err
	}
	err = containerBeginTemplate.Execute(w, nil)
	if err != nil {
		return err
	}

	// main context of this page
	err = tagcloudTemplate.Execute(w, ic)
	if err != nil {
		return err
	}

	err = tailTemplate.Execute(w, map[string]string{"footer": fmt.Sprintf("Version: %s", VERSION)})
	if err != nil {
		return err
	}
	return
}

func ErrorPage(w io.Writer, e error) (err error) {
	err = headTemplate.Execute(w, map[string]string{"title": "FileSrv :: ERROR"})
	if err != nil {
		return err
	}
	err = navbarTemplate.Execute(w, nil)
	if err != nil {
		return err
	}
	err = containerBeginTemplate.Execute(w, nil)
	if err != nil {
		return err
	}

	// main context of this page
	err = template.Must(template.New("serverError").Parse(`
  {{if .}}
<div class="span9">
  <div class="hero-unit">
    <h3>ERROR!</h3>
    <div class="error">
      {{.Error()}}
    </div>
  </div> {{/* hero-unit */}}
</div> {{/* span9 */}}
  {{end}}
  `)).Execute(w, e)
	if err != nil {
		return err
	}

	err = tailTemplate.Execute(w, map[string]string{"footer": fmt.Sprintf("Version: %s", VERSION)})
	if err != nil {
		return err
	}
	return
}
