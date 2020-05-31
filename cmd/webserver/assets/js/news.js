if ($('#news-page').length > 0) {

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

        const to = $(e.target);
        const from = $(e.relatedTarget);

        // On entering tab
        if (!to.attr('loaded')) {
            to.attr('loaded', 1);
            switch (to.attr('href')) {
                case '#all':
                    loadNewsAjax();
                    break;
            }
        }
    });

    // Show more article
    $('.minned').on('click', function (e) {
        $(this).toggleClass('minned');
    });

    function loadNewsAjax() {

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[3]);
                $(row).addClass('cursor-pointer');
            },
            "columnDefs": [
                // Article
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        const name = row[1] + '<br><small><a href="' + row[9] + '">' + row[7] + '</a></small>';
                        return '<div class="icon-name"><div class="icon"><img class="tall" alt="" src="" data-lazy="' + row[8] + '" data-lazy-alt="' + row[7] + '"></div><div class="name">' + name + '</div></div><div class="d-none">' + row[5] + '</div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false
                },
                // Date
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[10] + '" data-livestamp="' + row[5] + '"></span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderSequence": ['desc'],
                },
                // Search Score
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[6].toLocaleString();
                    },
                    "orderable": false,
                    "visible": false,
                },
            ]
        };

        const $table = $('#news-table');
        const table = $table.gdbTable({
            tableOptions: options,
            searchFields: [
                $('#search'),
            ],
        });

        $table.on('click', 'tr[role=row]', function () {

            const $tr = $(this);
            const row = table.row($tr);

            if (row.child.isShown()) {

                row.child.hide();
                $tr.removeClass('shown');

            } else {

                row.child($("<div/>").html(row.data()[2]).text()).show();
                $tr.addClass('shown');
            }
        });
    }
}
