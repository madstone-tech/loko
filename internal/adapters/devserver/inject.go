package devserver

import "bytes"

// reloadScript is added to HTML responses at serve time only. Built
// artifacts never contain it, so `loko build` output stays exactly what
// `serve` shows, minus this one element.
const reloadScript = `<script>
(function () {
  var es = new EventSource("/_loko/events");
  es.addEventListener("reload", function () { window.location.reload(); });
})();
</script>
`

// inject places the reload script just before </body>, or at the end when
// the page has none. It returns a new slice; the published bytes are never
// modified.
func inject(page []byte) []byte {
	i := bytes.LastIndex(page, []byte("</body>"))
	if i < 0 {
		return append(append([]byte(nil), page...), reloadScript...)
	}
	out := make([]byte, 0, len(page)+len(reloadScript))
	out = append(out, page[:i]...)
	out = append(out, reloadScript...)
	return append(out, page[i:]...)
}
