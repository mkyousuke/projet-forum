document.addEventListener("DOMContentLoaded", function() {
    const expandButtons = document.querySelectorAll(".btn-expand-overview");

    expandButtons.forEach(button => {
        button.addEventListener("click", function(e) {

            const filmContent = this.closest(".film-content");
            if (!filmContent) return;

            const overview = filmContent.querySelector(".film-overview");
            if (!overview) return;

            const isExpanded = overview.classList.toggle("expanded");

            this.textContent = isExpanded ? "Réduire" : "En savoir plus";
        });
    });
});