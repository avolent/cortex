/**
 * Dark mode toggle functionality
 * Persists preference to localStorage
 */

(function () {
  const darkModeToggle = document.getElementById('dark-mode-toggle');

  // Apply saved preference on page load
  if (localStorage.getItem('darkMode') === 'true') {
    document.body.classList.add('latex-dark');
  }

  // Toggle dark mode on button click
  if (darkModeToggle) {
    darkModeToggle.addEventListener('click', function () {
      document.body.classList.toggle('latex-dark');

      const isDarkMode = document.body.classList.contains('latex-dark');
      localStorage.setItem('darkMode', isDarkMode ? 'true' : 'false');
    });
  }
})();
