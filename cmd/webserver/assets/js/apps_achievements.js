const $appsAchievementsPage = $('#apps-achievements-page');

if ($appsAchievementsPage.length > 0) {

    const options = {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / App name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<a href="' + row[3] + '" class="icon-name"><div class="icon"><img alt="" data-lazy="' + row[2] + '" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Count
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Average
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[6].toLocaleString() + '%';
                },
                "orderSequence": ["desc"],
            },
            // Price
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[4];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Icons
            {
                "targets": 4,
                "render": function (data, type, row) {
                    if (isIterable(row[7])) {
                        return json2html.transform(row[7], {'<>': 'img', 'src': '', 'data-lazy': '${k}', 'data-lazy-alt': '${v}', 'class': 'mr-1', 'data-toggle': 'tooltip', 'data-placement': 'top', 'data-lazy-title': '${v}'});
                    }
                    return '';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
        ]
    };

    // Init table
    const searchFields = [
        $('#search'),
    ];

    $('table.table').gdbTable({
        tableOptions: options,
        searchFields: searchFields
    });
}
