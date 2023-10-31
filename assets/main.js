function isSearchClicked(clicked, targetBoxes) {
    do {
        if (targetBoxes.some(it => it === clicked)) { 
        document.getElementById("search-suggestion-box").hidden = false
            return true;
        }
        clicked = clicked.parentNode;
    } while (clicked);
    return false;
}


function clickHandler(evt) {
    const suggestion_box = document.getElementById("search-suggestion-box");
    const search_box = document.getElementById("search-bar");
    document.getElementById("search-suggestion-box").hidden = !isSearchClicked(evt.target, [suggestion_box, search_box]);
}

document.addEventListener("click", clickHandler);

