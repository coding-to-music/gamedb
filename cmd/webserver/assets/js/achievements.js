const $achievementsPage = $('#achievements-page');

if ($achievementsPage.length > 0) {

    const options = {
        "order": [[1, 'asc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[7]);
        },
        "columnDefs": [
            // Name
            {
                "targets": 0,
                "render": function (data, type, row) {

                    if (row[8]) {
                        row[2] = '<em>&lt;Hidden&gt;</em> ' + row[2];
                    }

                    let name = row[5] + ': ' + row[9] + '<br><small>' + row[2] + '</small>';

                    return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[1] + '" alt="" data-lazy-alt="' + row[0] + '"></div><div class="name">' + name + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Completed
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3] + '%';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    rowData[3] = Math.ceil(rowData[3]);
                    $(td).css('background', 'linear-gradient(to right, rgba(0,0,0,.15) ' + rowData[3] + '%, transparent ' + rowData[3] + '%)');
                    $(td).addClass('thin');
                },
                "orderSequence": ['desc', 'asc'],
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

    // Init table
    const searchFields = [
        $('#search'),
    ];

    $('table.table').gdbTable({
        tableOptions: options,
        searchFields: searchFields
    });
}
