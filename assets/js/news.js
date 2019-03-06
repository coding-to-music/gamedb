if ($('#news-page').length > 0) {

    const $modal = $('#news-modal');

    // Add hash when clicking row
    $('table.table').on('click', '.article-title', function (e) {
        history.pushState(undefined, undefined, '#apps,' + $(this).closest('tr').attr('data-id'));
        showArt();
    });

    // Remove hash when closing modal
    $modal.on('hidden.bs.modal', function (e) {
        history.pushState("", document.title, window.location.pathname + window.location.search + '#apps');
        showArt();
    });

    // News modal
    $(window).on('hashchange', showArt);
    $(document).on('draw.dt', showArt);

    function showArt() {

        const hash = window.location.hash.replace('#apps,', '');
        if (hash) {

            let $art = $('tr[data-id=' + hash + ']').find('.d-none').html();
            $art = $("<div />").html($art).text(); // Decode HTML
            $modal.find('.modal-body').html($art);
            $modal.modal('show');

        } else {
            $modal.modal('hide');
        }
    }

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

        const to = $(e.target);
        const from = $(e.relatedTarget);

        // On entering tab
        if (to.attr('href') === '#apps') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);

                loadNewsAjax();
            }
        }
    });

    function loadNewsAjax() {

        $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
            },
            "columnDefs": [
                // Game
                {
                    "targets": 0,
                    "render": function (data, type, row) {

                        // Icon URL
                        if (row[8]) {
                            row[8] = 'https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/' + row[6] + '/' + row[8] + '.jpg';
                        } else {
                            row[8] = '/assets/img/no-app-image-square.jpg';
                        }

                        return '<img src="' + row[8] + '" class="rounded square" alt="' + row[7] + '"><span data-app-id="' + row[6] + '">' + row[7] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                        $(td).attr('data-link', rowData[9]);
                    },
                    "orderable": false
                },
                // Title
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<span>' + row[1] + '</span><div class="d-none">' + row[5] + '</div>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('article-title');
                        $(td).attr('data-link', '');
                    },
                    "orderable": false
                },
                // Date
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '" data-livestamp="' + row[3] + '"></span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                }
            ]
        }));
    }

}
