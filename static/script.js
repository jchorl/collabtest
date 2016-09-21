fetch('/api/helloworld').then(function(resp) {
	return resp.text();
}).then(function(text) {
	document.getElementById('main').innerHTML = text;
});
