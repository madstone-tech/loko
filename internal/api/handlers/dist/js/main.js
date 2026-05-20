// Search functionality
document.addEventListener('DOMContentLoaded', function() {
	const searchInput = document.getElementById('search');
	if (searchInput) {
		searchInput.addEventListener('input', function(e) {
			const query = e.target.value.toLowerCase();
			const items = document.querySelectorAll('.system-link, .container-link');
			
			items.forEach(item => {
				const text = item.textContent.toLowerCase();
				const parent = item.closest('li');
				if (text.includes(query)) {
					parent.style.display = '';
				} else {
					parent.style.display = 'none';
				}
			});
		});
	}
});

// Smooth scroll for anchor links
document.addEventListener('click', function(e) {
	if (e.target.tagName === 'A' && e.target.getAttribute('href').startsWith('#')) {
		e.preventDefault();
		const target = document.querySelector(e.target.getAttribute('href'));
		if (target) {
			target.scrollIntoView({ behavior: 'smooth' });
		}
	}
});

// Back to top button
window.addEventListener('scroll', function() {
	const backToTop = document.querySelector('.back-to-top');
	if (backToTop) {
		if (window.scrollY > 300) {
			backToTop.style.display = 'block';
		} else {
			backToTop.style.display = 'none';
		}
	}
});

// Highlight active navigation item
document.addEventListener('DOMContentLoaded', function() {
	const currentPath = window.location.pathname;
	const navLinks = document.querySelectorAll('.system-link, .container-link');
	
	navLinks.forEach(link => {
		const href = link.getAttribute('href');
		if (href && currentPath.includes(href.replace(/^\//, ''))) {
			link.closest('li').classList.add('active');
		}
	});
});