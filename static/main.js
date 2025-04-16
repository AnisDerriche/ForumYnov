document.addEventListener("DOMContentLoaded", () => {
    const path = window.location.pathname;
    let category = "all";

    if (path.includes("cyber")) {
        category = "cyber";
    } else if (path.includes("info")) {
        category = "info";
    } else if (path.includes("anglais")) {
        category = "anglais";
    }

    fetchPosts(category);
});

function fetchPosts(category = "all") {
    let url = "/posts";
    if (category !== "all") {
        url = `/posts/${category}`;
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

                    <div class="comments" id="comments-${post.id}"></div>

                    <form class="comment-form" data-post-id="${post.id}">
                        <input type="email" name="email" placeholder="Ton email" required>
                        <input type="text" name="contenu" placeholder="Ton commentaire" required>
                        <button type="submit">Envoyer</button>
                    </form>
                `;

                container.appendChild(postDiv);

                // Afficher les commentaires existants
                fetch(`/comments/${post.id}`)
                    .then(res => res.json())
                    .then(comments => {
                        const commentsDiv = document.getElementById(`comments-${post.id}`);
                        if (comments.length === 0) {
                            commentsDiv.innerHTML = "<p>Aucun commentaire pour ce post.</p>";
                        } else {
                            comments.forEach(comment => {
                                const commentHTML = `
                                    <div class="comment">
                                        <p>${sanitize(comment.contenu)}</p>
                                        <small>Par ${sanitize(comment.email)} - ${formatDate(comment.created_at)}</small>
                                    </div>
                                `;
                                commentsDiv.innerHTML += commentHTML;
                            });
                        }
                    });

                // Gérer l'envoi du commentaire
                postDiv.querySelector(".comment-form").addEventListener("submit", function (e) {
                    e.preventDefault();

                    const form = e.target;
                    const postID = form.getAttribute("data-post-id");
                    const email = form.email.value;
                    const contenu = form.contenu.value;

                    fetch(`/comments`, {
                        method: "POST",
                        headers: {
                            "Content-Type": "application/json"
                        },
                        body: JSON.stringify({
                            post_id: postID,
                            email: email,
                            contenu: contenu
                        })
                    })
                    .then(res => {
                        if (!res.ok) throw new Error("Erreur lors de l'ajout du commentaire");
                        return res.json();
                    })
                    .then(() => {
                        form.reset();
                        fetchPosts(category); // Recharge les posts pour afficher le nouveau commentaire
                    })
                    .catch(err => {
                        console.error(err);
                        alert("Erreur lors de l'envoi du commentaire.");
                    });
                });
            });
        })
        .catch((err) => {
            console.error("Erreur lors de la récupération des posts :", err);
        });
}

function setupCommentFormListeners() {
    document.querySelectorAll(".comment-form").forEach((form) => {
        form.addEventListener("submit", (e) => {
            e.preventDefault();

            const postID = form.getAttribute("data-post-id");
            const contenu = form.querySelector("input[name='contenu']").value;

            fetch(`/comments`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    post_id: postID,
                    email: email,
                    contenu: contenu
                })
            })
            .then((res) => {
                if (!res.ok) {
                    throw new Error("Erreur lors de l'ajout du commentaire");
                }
                return res.json();  // Vérifiez que la réponse est bien JSON
            })
            .then((data) => {
                console.log("Réponse après ajout du commentaire:", data);  // Log pour déboguer
                form.reset();
                fetchPosts(category); // Recharge les posts pour afficher le nouveau commentaire
            })
            .catch((err) => {
                console.error("Erreur lors de l'envoi du commentaire :", err);
                alert("Erreur lors de l'envoi du commentaire.");
            });
            
        });
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
