package main

import (
	"github.com/russross/blackfriday"
)

const (
	//https://github.com/russross/blackfriday/blob/master/markdown.go#L32
	extensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		//	githubの拡張マークダウン同様の改行
		// translate newlines into line breaks
		blackfriday.EXTENSION_HARD_LINE_BREAK |
		blackfriday.EXTENSION_DEFINITION_LISTS |
		0 |
		blackfriday.EXTENSION_FOOTNOTES // Pandoc-style footnotes

	markdownTemplateHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>%s</title>
<link rel="stylesheet" href="/_assets/sanitize.min.css" media="all">
<link rel="stylesheet" href="/_assets/bootstrap.min.css">
<link rel="stylesheet" href="/_assets/github-markdown.css" media="all">
<link rel="stylesheet" href="/_assets/style.css" media="all">

<!-- for offline -->
<link rel="stylesheet" href="/_assets/sons-of-obsidian.css" media="all">

<script src="/_assets/run_prettify.js?autoload=true&amp;skin=sons-of-obsidian" defer="defer"></script>

<!-- <script type="text/javascript" src="/_assets/MathJax.js?config=TeX-AMS-MML_HTMLorMML"></script> -->
<script type="text/javascript" src="http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS-MML_HTMLorMML"></script>

<script type="text/x-mathjax-config">
	MathJax.Hub.Config({ tex2jax: { inlineMath: [['$','$'], ["\\(","\\)"]] } });
</script>

<script src="/_assets/jquery-2.1.1.min.js"></script>
<script>
$(function() {
	//	e.g.
	//	class="language:go-main.go"
	//	---> class="language:go" <h4>main.go</h4>
	$('pre>code').each(function() {
		var e = $(this);
		var p = $(this.parentNode);
		var tmp = e.attr("class");
		if (tmp !== undefined) {
			$.each(String(tmp).split(' '), function(i, val) {
			var list = val.match(/^(language-.*):(.*)$/)
			if (list !== null) {
				e.removeClass(val);
				e.addClass(list[1]);
				p.before('<span class="code_file_name">'+list[2]+'</span>');
			}
			});
		}
		p.addClass('prettyprint').addClass('linenums');
	});
	//	google/code-prettifyではない場合?!
	//prettyPrint();
	$.getScript(window.location.protocol + '//' + window.location.hostname + '%s/livereload.js');

	//	var-start ~ var-end の要素をdata-target-idで指定した箇所にinsert
	var starts = $('.var-start');
	var ends   = $('.var-end');
	if (starts.length != ends.length) {
		console.log(".var-start length != .var-end length : ", starts.length, " != ", ends.length);
	} else {
		starts.each(function(i, v) {
			var start = $(v);
			var end = $(ends[i]);
			var content = start.nextUntil(end);
			var target_id = '#'+start.data('target-id');
			content.appendTo(target_id);
		});
	}
	});
</script>
</head>
<body>

<div style="margin: 10px;">
	<a href="/">Home</a>
</div>
<hr size="4">
<br>

<div class="markdown-body">%s</div>

<br>
<hr size="4">
<div style="margin: 10px;">
	<a href="/">Home</a>
</div>
</body>
</html>
`

	goexecTemplateHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>%s</title>
<link rel="stylesheet" href="/_assets/sanitize.css" media="all">
<link rel="stylesheet" href="/_assets/github-markdown.css" media="all">
<link rel="stylesheet" href="/_assets/sons-of-obsidian.css" media="all">
<link rel="stylesheet" href="/_assets/style.css" media="all">

<script type="text/javascript" src="http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS-MML_HTMLorMML"></script>
<script type="text/x-mathjax-config">
	MathJax.Hub.Config({ tex2jax: { inlineMath: [['$','$'], ["\\(","\\)"]] } });
</script>
<!-- <script src="/_assets/MathJax.js"></script> -->

<script src="/_assets/jquery-2.1.1.min.js"></script>
<script>
$(function() {
	$.getScript(window.location.protocol + '//' + window.location.hostname + '%s/livereload.js');
});
</script>
</head>
<body>

<div style="margin: 10px;">
	<a href="/">Home</a>
</div>
<hr size="4">
<br>

<div>%s</div>

<br>
<hr size="4">
<div style="margin: 10px;">
	<a href="/">Home</a>
</div>

</body>
</html>
`
)

const dirTreeTemplateHTML = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8" />
	<title>{{.Title}}</title>

	<link rel="stylesheet" href="/_assets/bootstrap.min.css">
	<link rel="stylesheet" type="text/css" href="/_assets/tree.css">
	<link rel="stylesheet" type="text/css" href="/_assets/app.css">
</head>
<body>
	<h1>{{.Title}}</h1>
	<!--
	<ul class="tree-menu">
		{{.Links}}
	</ul>
	-->

	<script async>
	// [HTML Imports: ウェブのための #include - HTML5 Rocks]( https://www.html5rocks.com/ja/tutorials/webcomponents/imports/ )
	function handleLoad(e) {
		var link = document.querySelector('link[rel="import"]');
		var content = link.import;

		var el = content.querySelector('.extended_html');

		document.body.appendChild(el.cloneNode(true));

		console.log('Loaded import: ' + e.target.href);
	}
	function handleError(e) {
		console.log('Error loading import: ' + e.target.href);
	}

	function supportsImports() {
		return 'import' in document.createElement('link');
	}

	if (supportsImports()) {
		// Good to go!
	} else {
		// Use other libraries/require systems to load files.
	}
	</script>

<!--	<link rel="import" href="/_assets/vue-app.html" onload="handleLoad(event)" onerror="handleError(event)"> -->

	<script src="/_assets/jquery-2.1.1.min.js"></script>
	<script src="/_assets/jquery.mark.min.js"></script>
	<script src="/_assets/vue.js"></script>
	<script src="/_assets/app.js"></script>
</body>
</html>
`
