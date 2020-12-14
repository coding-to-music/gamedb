if ($('#news-page').length > 0) {

    // Setup drop downs
    $('select.form-control-chosen').chosen({
        disable_search_threshold: 5,
        allow_single_deselect: true,
        max_selected_options: 10
    });

    const $table = $('#news-table');
    const table = $table.gdbTable({
        searchFields: [
            $('#filter'),
            $('#feed'),
            $('#search'),
        ],
        tableOptions: {
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
                        const name = row[10] + '<br><small><a href="' + row[8] + '">' + row[7] + '</a> - ' + row[11] + '</small>';
                        return '<div class="icon-name"><div class="icon"><img class="tall" alt="" src="" data-lazy="' + row[4] + '" data-lazy-alt="' + row[1] + '"></div><div class="name">' + name + '</div></div>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Date
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" data-livestamp="' + row[5] + '"></span>'
                            + '<br><small class="text-muted">' + row[9] + '</small>';
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
                        return row[6];
                    },
                    "orderable": false,
                    "visible": user.isLocal,
                },
            ]
        },
    });

    $table.on('click', 'tbody tr[role=row]', function () {

        const $tr = $(this);
        const row = table.row($tr);

        if (row.child.isShown()) {

            row.child.hide();
            $tr.removeClass('shown');

        } else {
            row.child($("<div/>").html(row.data()[2])).show();
            $tr.addClass('shown');
            observeLazyImages($(this).next().find('img[data-lazy]'));
        }
    });

    $table.on('click', 'a', function (e) {
        e.stopPropagation();
    });

    websocketListener('news', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = JSON.parse(e.data);
            const type = $('#filter').val();

            if (type === '') {

                addDataTablesRow(options, data.Data, info.length, $table);

                console.log(data.Data);
                return

                // Sort rows
                const rows = $table.find('tr').get();
                rows.sort(function (a, b) {
                    const keyA = $(a).attr('myAttribute');
                    const keyB = $(b).attr('myAttribute');

                    return Math.sign(keyA - keyB);
                });

                $.each(rows, function (index, row) {
                    $table.children('tbody').append(row);
                });
            }
        }
    });
}
