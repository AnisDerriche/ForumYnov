function filterArtists() {
    let input = document.getElementById("searchBar").value.toLowerCase();
    let artists = document.getElementsByClassName("artist-card");

    for (let i = 0; i < artists.length; i++) {
        let name = artists[i].getElementsByTagName("h2")[0].innerText.toLowerCase();
        if (name.includes(input)) {
            artists[i].style.display = "";
        } else {
            artists[i].style.display = "none";
        }
    }
}
