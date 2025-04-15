document.addEventListener("DOMContentLoaded", () => {
    fetch("/posts")
        .then(response => response.json())
        .then(posts => {
            const container = document.getElementById("posts-container");
            if (!container) return;

            posts.forEach(post => {
                const postElement = document.createElement("div");
                postElement.classList.add("post");

                postElement.innerHTML = `
                    <h2>${post.title}</h2>
                    <p><strong>${post.email}</strong> - <em>${new Date(post.created_at).toLocaleString()}</em></p>
                    <p>${post.contenu}</p>
                    <hr>
                `;

                container.appendChild(postElement);
            });
        })
        .catch(error => {
            console.error("Erreur lors de la récupération des posts :", error);
        });
});

document.addEventListener("DOMContentLoaded", () => {
    fetchPosts();
});

function fetchPosts() {
    fetch("/posts")
        .then((res) => res.json())
        .then((posts) => {
            const container = document.getElementById("posts-container");
            container.innerHTML = "";

            if (posts.length === 0) {
                container.innerHTML = "<p>Aucun post pour le moment. Sois le premier à poster ! ✨</p>";
                return;
            }

            posts.forEach((post) => {
                const postDiv = document.createElement("div");
                postDiv.className = "post";

                postDiv.innerHTML = `
                    <div class="post-title">${sanitize(post.title)}</div>
                    <div class="post-email">Posté par : ${sanitize(post.email)}</div>
                    <div class="post-contenu">${sanitize(post.contenu)}</div>
                    <div class="post-date">${formatDate(post.created_at)}</div>
                `;

                container.appendChild(postDiv);
            });
        })
        .catch((err) => {
            console.error("Erreur lors de la récupération des posts :", err);
        });
}

// Sécurité basique contre l'injection HTML
function sanitize(str) {
    const temp = document.createElement("div");
    temp.textContent = str;
    return temp.innerHTML;
}

// Format date à la française
function formatDate(isoString) {
    const date = new Date(isoString);
    return date.toLocaleString("fr-FR", {
        dateStyle: "short",
        timeStyle: "short",
    });
}

// main.js

document.addEventListener('DOMContentLoaded', function() {
    const form = document.querySelector('form');
    const successMessage = document.createElement('div');
    const errorMessage = document.createElement('div');
    
    form.addEventListener('submit', function(event) {
        event.preventDefault();

        const title = document.getElementById('title').value;
        const email = document.getElementById('email').value;
        const content = document.getElementById('contenu').value;

        if (title && email && content) {
            // Simulate form submission success
            showSuccessMessage("Le post a été créé avec succès !");
            
            // You can implement your actual submission logic here (AJAX or redirect)
            setTimeout(() => {
                window.location.href = "/"; // Redirect to the homepage after 2 seconds
            }, 2000);
        } else {
            showErrorMessage("Tous les champs doivent être remplis !");
        }
    });

    function showSuccessMessage(message) {
        successMessage.textContent = message;
        successMessage.classList.add('success');
        form.appendChild(successMessage);
        successMessage.style.display = 'block';
        errorMessage.style.display = 'none';
    }

    function showErrorMessage(message) {
        errorMessage.textContent = message;
        errorMessage.classList.add('error');
        form.appendChild(errorMessage);
        errorMessage.style.display = 'block';
        successMessage.style.display = 'none';
    }
});
