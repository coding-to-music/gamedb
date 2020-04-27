if ($('#groups-page').length > 0) {

    const options = {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-group-id', data[0]);
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Rank
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return row[11].toLocaleString();
                },
                "orderable": false,
            },
            // Icon / Name
            {
                "targets": 1,
                "render": function (data, type, row) {

                    let name = row[1];
                    if (row[9]) {
                        name += '<span class="badge badge-danger float-right">Removed</span>';
                    }

                    return '<div class="icon-name"><div class="icon"><img data-src="/assets/img/no-app-image-square.jpg" data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + name + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Members
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Trend Value
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[10].toLocaleString();
                },
                "orderSequence": ["desc", "asc"],
            },
            // Link
            {
                "targets": 4,
                "render": function (data, type, row) {
                    if (row[8]) {
                        return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
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
            // $('#type'),
            // $('#errors'),
        ],
    });
}
