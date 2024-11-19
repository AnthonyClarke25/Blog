document.querySelectorAll('.post a').forEach(postLink => {
    postLink.addEventListener('click', () => {
        alert("You clicked on a blog post!");
    });
});