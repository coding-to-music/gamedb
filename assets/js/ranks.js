if ($('#ranks-page').length > 0) {

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[3, 'asc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[1]);
            $(row).attr('data-link', '/players/' + data[1]);
        },
        "columnDefs": [
            // Rank
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return row[0];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('font-weight-bold')
                },
                "orderable": false
            },
            // Player
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<img src="' + row[3] + '" class="rounded square"><span>' + row[2] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                },
                "orderable": false
            },
            // Flag
            {
                "targets": 2,
                "render": function (data, type, row) {
                    if (row[11]) {
                        return '<img data-toggle="tooltip" data-placement="left" title="' + row[12] + '" src="' + row[11] + '" class="rounded">';
                    }
                    return '';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false
            },
            // Avatar 2 / Level
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '<div class="' + row[4] + ' square"></div><span>' + row[5].toLocaleString() + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                }
            },
            // Games
            {
                "targets": 4,
                "render": function (data, type, row) {

                    if (row[6]) {
                        return row[6].toLocaleString();
                    }
                    return $lockIcon;
                }
            },
            // Badges
            {
                "targets": 5,
                "render": function (data, type, row) {
                    return row[7].toLocaleString();
                }
            },
            // Time
            {
                "targets": 6,
                "render": function (data, type, row) {

                    if (row[8] === '0m') {
                        return $lockIcon;
                    }

                    return row[8];
                },
                "createdCell": function (td, cellData, rowData, row, col) {

                    $(td).attr('nowrap', 'nowrap');

                    if (rowData[8] !== '0m') {
                        $(td).attr('data-toggle', 'tooltip').attr('data-placement', 'left').attr('title', rowData[9]);
                    }
                }
            },
            // Friends
            {
                "targets": 7,
                "render": function (data, type, row) {
                    return row[10].toLocaleString();
                }
            }
        ]
    }));
}