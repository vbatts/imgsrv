package main

import (
  "html/template"
  "io"
)

var index = `
<html>
<head>
</head>
<body>


</body>
</html>
`
func IndexPage(w io.Writer) (err error) {
  t, err := template.New("index").Parse(index)
  if (err != nil) {
    return err
  }
  return t.Execute(w, index)
}

