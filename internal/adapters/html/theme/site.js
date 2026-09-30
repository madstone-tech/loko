// Client-side search over the element index in search-index.js, and
// highlighting of the current page in the navigation. No network access.
(function () {
  "use strict";
  document.addEventListener("DOMContentLoaded", function () {
    var nav = document.querySelector(".sidebar-nav");
    var root = nav ? nav.getAttribute("data-root") || "" : "";
    var input = document.getElementById("search");
    var results = document.getElementById("search-results");
    var index = window.LOKO_SEARCH || [];

    if (input && results) {
      input.addEventListener("input", function () {
        var q = input.value.trim().toLowerCase();
        results.innerHTML = "";
        results.hidden = q === "";
        if (q === "") return;
        index.filter(function (e) {
          return (e.n + " " + e.k + " " + e.d).toLowerCase().indexOf(q) >= 0;
        }).slice(0, 50).forEach(function (e) {
          var li = document.createElement("li");
          var a = document.createElement("a");
          a.href = root + e.p;
          a.textContent = e.n;
          var small = document.createElement("small");
          small.textContent = " " + e.k;
          li.appendChild(a);
          li.appendChild(small);
          results.appendChild(li);
        });
      });
    }

    var here = window.location.pathname;
    document.querySelectorAll(".system-link").forEach(function (link) {
      var href = link.getAttribute("href") || "";
      if (href && here.slice(-href.replace(/^(\.\.\/)+/, "").length) === href.replace(/^(\.\.\/)+/, "")) {
        link.closest("li").classList.add("active");
      }
    });
  });
})();
