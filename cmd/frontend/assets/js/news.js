if ($('#news-page').length > 0) {

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
    };

    const $table = $('#news-table');
    const table = $table.gdbTable({
        tableOptions: options,
        searchFields: [
            $('#filter'),
            $('#feed'),
            $('#search'),
        ],
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
}
