const RESULT_LIMIT = 10;
const Controller = {
    search: (ev) => {
        ev.preventDefault();
        const form = document.getElementById("form");
        const data = Object.fromEntries(new FormData(form));
        let page = '1'
        if (ev.target.tagName.toLowerCase() === 'a') {
            page = ev.target.innerHTML;
        }

        const query = data.query
        const response = fetch(`/search?q=${query}&limit=10&page=${page}`).then((response) => {
            response.json().then((response) => {
                Controller.updateTable(response);
            });
        });
    },

    updateTable: (response) => {
        let table = document.getElementById("table-body");
        table.innerHTML = "";

        for (let work of response.works) {
            for (let result of work.results) {
                let tr = table.insertRow();
                let td = document.createElement('td');
                td.innerHTML = work.title;
                td.style.fontWeight = "bold";
                td.style.fontSize = "18px";
                tr.appendChild(td);

                tr = table.insertRow();
                td = document.createElement('td');
                td.innerHTML = result;
                tr.appendChild(td);
            }
        }
    },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);