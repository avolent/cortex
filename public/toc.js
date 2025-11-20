/**
 * Table of Contents organization
 * Moves the contents heading and list into a navigation wrapper
 */

(function () {
  const contents = document.getElementById('contents');

  if (contents && contents.nextElementSibling) {
    const list = contents.nextElementSibling;

    // Create navigation wrapper
    const nav = document.createElement('nav');
    nav.id = 'nav';
    nav.setAttribute('role', 'navigation');
    nav.className = 'toc';

    // Insert nav before contents
    contents.parentNode.insertBefore(nav, contents);

    // Move contents and list into nav
    nav.appendChild(contents);
    nav.appendChild(list);
  }
})();
