if ($('#news-page').length > 0) {

    const $table = $('table.table-datatable2');

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

    // Show more article
    $('.minned').on('click', function (e) {
        $(this).toggleClass('minned');
    });

    function loadNewsAjax() {

        const table = $table.DataTable($.extend(true, {}, dtDefaultOptions, {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[6]);
            },
            "columnDefs": [
                // Game
                {
                    "targets": 0,
                    "render": function (data, type, row) {

                        // Icon URL
                        if (row[8] === '') {
                            row[8] = '/assets/img/no-app-image-square.jpg';
                        } else if (!row[8].startsWith("/") && !row[8].startsWith("http")) {
                            row[8] = 'https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/' + row[6] + '/' + row[8] + '.jpg';
                        }

                        return '<img src="' + row[8] + '" class="rounded square" alt="' + row[7] + '"><span>' + row[7] + '</span>';
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

        $table.on('click', 'td.article-title', function () {

            const $tr = $(this).closest('tr');
            const row = table.row($tr);

            if (row.child.isShown()) {

                row.child.hide();
                $tr.removeClass('shown');

            } else {

                row.child($("<div/>").html(row.data()[5]).text()).show();
                $tr.addClass('shown');
            }
        });
    }
}
