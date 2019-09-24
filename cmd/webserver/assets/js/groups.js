if ($('#groups-page').length > 0) {

    const options = {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-group-id64', data[0]);
            $(row).attr('data-group-id', data[11]);
            $(row).attr('data-link', data[2]);
            if (data[7] === 'game' && !$('#type').val()) {
                $(row).addClass('table-primary');
            }
            if (data[9]) {
                $(row).addClass('table-danger');
            }
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-src="/assets/img/no-app-image-square.jpg" data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Members
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Trend Value
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[10].toLocaleString();
                },
                "orderSequence": ["desc", "asc"],
            },
            // Link
            {
                "targets": 3,
                "render": function (data, type, row) {
                    if (row[8]) {
                        return '<a href="' + row[8] + '" target="_blank" rel="nofollow"><i class="fas fa-link" data-target="_blank"></i></a>';
                    }
                    return '';
                },
                "orderable": false,
            },
        ]
    };

    $('table.table').gdbTable({
        tableOptions: options,
        searchFields: [
            $('#search'),
            $('#type'),
            $('#errors'),
        ],
    });
}
