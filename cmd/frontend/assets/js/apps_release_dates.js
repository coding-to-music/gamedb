if ($('#release-dates-page').length > 0) {

    $('table.table').gdbTable({
        tableOptions: {
            "order": [[4, 'asc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[3]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<a href="' + row[3] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
                // Release Date
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    "orderable": false,
                },
                // Followers
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[4].toLocaleString();
                    },
                    "orderSequence": ["desc"],
                },
                // External Link
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        if (row[5]) {
                            return '<a href="' + row[5] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
                // Search Score
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        return row[7];
                    },
                    "visible": user.isLocal,
                    "orderable": false,
                    "orderSequence": ["asc", "desc"],
                },
            ]
        },
    });
}
