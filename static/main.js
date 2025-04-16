document.addEventListener("DOMContentLoaded", () => {
    // Déterminer la catégorie à partir du nom de la page
    const path = window.location.pathname;
    let category = "all";

    if (path.includes("cyber")) {
        category = "cyber";
    } else if (path.includes("info")) {
        category = "info";
    } else if (path.includes("anglais")) {
        category = "anglais";
    }

    // Charger les posts selon la catégorie détectée
    fetchPosts(category);
});

function fetchPosts(category = "all") {
    let url = "/posts";
    if (category !== "all") {
        url = `/posts/${category}`; // <-- Correction ici
    }

    fetch(url)
        .then((res) => res.json())
        .then((posts) => {
            const container = document.getElementById("posts-container");
            container.innerHTML = "";

            if (posts.length === 0) {
                container.innerHTML = "<p>Aucun post pour le moment dans cette catégorie. Sois le premier à poster ! ✨</p>";
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

function sanitize(str) {
    const temp = document.createElement("div");
    temp.textContent = str;
    return temp.innerHTML;
}

function formatDate(isoString) {
    const date = new Date(isoString);
    return date.toLocaleString("fr-FR", {
        dateStyle: "short",
        timeStyle: "short",
    });
}
