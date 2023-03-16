const RESULT_LIMIT = 10;
const Controller = {
    search: (ev) => {
        ev.preventDefault();
        const form = document.getElementById("form");
        const data = Object.fromEntries(new FormData(form));
        const query = data.query
        const response = fetch(`/search?q=${query}`).then((response) => {
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
                // title
                let tr = table.insertRow();
                let td = document.createElement('td');
                td.style.fontWeight = "bold";
                td.style.fontSize = "18px";
                tr.appendChild(td);
                let a = document.createElement('a');
                a.innerHTML = work.title;
                td.appendChild(a);
                a.onclick = function(ev) {
                    Controller.getWork(encodeURI(ev.target.text));
                }

                // results
                tr = table.insertRow();
                td = document.createElement('td');
                td.innerHTML = result;
                tr.appendChild(td);
            }
        }

        table.hidden = false;
        document.getElementById("work").hidden = true;
    },

    getWork: (title) => {
        const response = fetch(`/work?t=${title}`).then((response) => {
            response.json().then((response) => {
                if (response.hasOwnProperty("title")) {
                    Controller.updateWorkContent(response);
                } else {
                    console.log("response", response);
                    console.log("missing work title");
                    let errResponse = { title: "missing work title", contents: "" }
                    Controller.updateWorkContent(errResponse);
                }
            });
        });
    },

    updateWorkContent: (response) => {
        document.getElementById("table-body").hidden = true;
        document.getElementById("title").innerText = response['title'];
        document.getElementById("contents").innerText = response['contents'];
        document.getElementById("work").hidden = false;
    }
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);