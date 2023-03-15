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
                Controller.updatePagination(response, query);
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
    updatePagination: (response, query) => {
        let pagination = document.getElementById("pagination");
        pagination.innerHTML = "";
        let previous = false;

        for (let page = response.page, start = page, end = page + RESULT_LIMIT; page <= end && page < response.totalPages; page++) {
            previousPage = start - 1
            if (previous == false && previousPage > 0) {
                paginate(pagination, previousPage);
                previous = true;
            }

            if (page == start) {
                let firstPage = "";
                if (page == 1) {
                    firstPage = "disabled";
                }
                paginate(pagination, page, firstPage);
            }

            if (page > response.page && page < end) {
                paginate(pagination, page);
            }
        }

        function paginate(pagination, page, disabled = '') {
            var li = document.createElement("li");
            li.className = "page-item " + disabled;
            var a = document.createElement("a");
            a.className = "page-link ";
            a.setAttribute("href", `/search?q=${query}&page=${page}`);
            a.innerHTML = page;
            li.appendChild(a);
            pagination.appendChild(li);
            return { li, a };
        }
    },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
document.getElementById("pagination").addEventListener("click", Controller.search);