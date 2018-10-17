package main

import (
	"encoding/json"
	"fmt"
	"github.com/devplayg/ipas-server"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	printHeader(w)
	html := `
<div class="row" style="padding: 20px;">
	<div class="col-sm-6" >
		<div class="alert alert-secondary" role="alert">
            Request (JSON type)
		</div>
		<form id="form-json">
			<div class="form-group">
    			<textarea class="form-control" name="data" rows="5">{"key":"valid_value"}</textarea>
  			</div>
			<button id="btn-send" class="btn btn-primary">Request</button>
		</form>
	</div>
	<div class="col-sm-6">
		<div class="alert alert-primary" role="alert">
  			Output
		</div>
		<pre class="result" style="padding: 5px; border: 1px dashed #acacac"></pre>
	</div>
</div>

	`
	fmt.Fprint(w, html)
	printFooter(w)
}

func ParseData(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	//spew.Dump(r.Body)
	//fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))

	decoder := json.NewDecoder(req.Body)
	jsonMap := make(map[string]interface{})
	err := decoder.Decode(&jsonMap)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	jsonRes, err := json.MarshalIndent(jsonMap, "", "    ")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	} else {
		fmt.Fprintf(w, string(jsonRes))
	}
}

func main() {
	var (
		port = ipasserver.CmdFlags.String("p", ":80", "HTTP Port")
	)
	ipasserver.CmdFlags.Usage = ipasserver.PrintHelp
	ipasserver.CmdFlags.Parse(os.Args[1:])

	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/event", ParseData)

	log.Fatal(http.ListenAndServe(*port, router))
}

func printHeader(w http.ResponseWriter) {
	html := `
<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css" integrity="sha384-MCw98/SFnGE8fJT3GXwEOngsV7Zt27NXFoaoApmYm81iuXoPkFOJwJ8ERdknLPMO" crossorigin="anonymous">
    <title></title>
  </head>
  <body>
`
	fmt.Fprint(w, html)
}

func printFooter(w http.ResponseWriter) {
	html := `
	<script src="https://code.jquery.com/jquery-3.1.1.min.js">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.3/umd/popper.min.js" integrity="sha384-ZMP7rVo3mIykV+2+9J3UJ46jBk0WLaUAdn689aCwoqbBJiSnjAK/l8WvCWPIPm49" crossorigin="anonymous"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/js/bootstrap.min.js" integrity="sha384-ChfqqxuZUCnJSK3+MXmPNIyE6ZbWh2IMqE241rYiqJxyMiZ6OW/JmZQ5stwEULTy" crossorigin="anonymous"></script>
    <script>
		$( document ).ready(function() {
			$( "#btn-send" ).click(function( e ) {
				e.preventDefault();
				$.ajax({
					method: "POST",
					url: "/event",
					data: $("#form-json textarea[name=data]").val()
				}).done(function( msg ) {
					$( ".result" ).html( msg );
				}).fail(function() {
					$( ".result" ).html( "failed to request" );
				});
			});
		});
	</script>
  </body>
</html>
`
	fmt.Fprint(w, html)
}
