if ($('#sales-page').length > 0) {

    (function ($, document) {
        'use strict';

        // Count down
        const $clock = $('#clock');
        const timestamp = $clock.attr('data-time');

        $clock.countdown(timestamp, function (e) {
            $(this).html(e.strftime('%D:%H:%M:%S'));
        });

        // Setup drop downs
        $('select.form-control-chosen').chosen({
            disable_search_threshold: 5,
            allow_single_deselect: true,
            max_selected_options: 10,
        });

        // Price slider
        const $priceElement = $('#price-slider');
        const priceSlider = noUiSlider.create($priceElement[0], {
            start: [0, 100],
            connect: true,
            step: 1,
            range: {
                'min': 0,
                'max': 100,
            },
        });

        // Score slider
        const $scoreElement = $('#score-slider');
        const scoreSlider = noUiSlider.create($scoreElement[0], {
            start: [0, 100],
            connect: true,
            step: 1,
            range: {
                'min': 0,
                'max': 100,
            },
        });

        // Discount slider
        const $discountElement = $('#discount-slider');
        const discountSlider = noUiSlider.create($discountElement[0], {
            start: [0, 100],
            connect: true,
            step: 1,
            range: {
                'min': 0,
                'max': 100,
            },
        });

        // Index slider
        const $indexElement = $('#index-slider');
        const indexMax = parseInt($indexElement.attr('data-max'));
        const indexSlider = noUiSlider.create($indexElement[0], {
            start: indexMax + 1,
            connect: true,
            step: 1,
            range: {
                'min': 1,
                'max': indexMax + 1,
            },
        });

        //
        function updateLabels(e) {

            const prices = priceSlider.get();
            if (prices[0] === prices[1]) {
                $('label#price-label').html('Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ')');
            } else {
                $('label#price-label').html('Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ' to ' + user.userCurrencySymbol + Math.round(prices[1]) + ')');
            }

            const scores = scoreSlider.get();
            if (scores[0] === scores[1]) {
                $('label#score-label').html('Score (' + Math.round(scores[0]) + '%)');
            } else {
                $('label#score-label').html('Score (' + Math.round(scores[0]) + '% to ' + Math.round(scores[1]) + '%)');
            }

            const discounts = discountSlider.get();
            if (discounts[0] === discounts[1]) {
                $('label#discount-label').html('Discount (-' + Math.round(discounts[0]) + '%)');
            } else {
                $('label#discount-label').html('Discount (-' + Math.round(discounts[0]) + '% - to -' + Math.round(discounts[1]) + '%)');
            }

            $('label#index-label').html('Max Per Game (' + Math.trunc(indexSlider.get()) + ')');
        }

        window.updateLabels = updateLabels;

        $(updateLabels);

        //
        const options = {
            'order': [[3, 'desc']],
            'createdRow': function (row, data, dataIndex) {
                $(row).attr('data-link', data[3]);
                $(row).attr('data-app-id', data[0]);
            },
            'columnDefs': [
                // Icon / App Name
                {
                    'targets': 0,
                    'render': function (data, type, row) {

                        let field = row[1];
                        field = field + ' <br /><small>' + row[13] + ' / ' + row[10] + '</small>';

                        if (row[11] === 1) {
                            field = field + ' <span class="badge badge-success float-right">Equal lowest</span>';
                        } else if (row[11] === 2) {
                            field = field + ' <span class="badge badge-success float-right">Lowest Ever!</span>';
                        }

                        if (row[15] != null && row[15].includes(777)) {
                            field = field + ' <span class="badge badge-success float-right">Low Confidence</span>';
                        }

                        return '<a href="' + row[3] + '" class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + field + '</div></a>';
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    // "orderable": false,
                },
                // Price
                {
                    'targets': 1,
                    'render': function (data, type, row) {
                        if (!row[4]) {
                            return '-';
                        }
                        return row[4];
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // Discount
                {
                    'targets': 2,
                    'render': function (data, type, row) {
                        return row[5] + '%';
                    },
                    'orderSequence': ['asc'],
                },
                // Rating
                {
                    'targets': 3,
                    'render': function (data, type, row) {
                        return row[6];
                    },
                    'orderSequence': ['desc'],
                },
                // End Date
                {
                    'targets': 4,
                    'render': function (data, type, row) {

                        let time = '<span data-toggle="tooltip" data-placement="left" title="' + row[7] + '" data-livestamp="' + row[7] + '"></span>';
                        if (row[11]) {
                            time = time + '*';
                        }
                        return time;

                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['asc'],
                },
                // Release Date
                {
                    'targets': 5,
                    'render': function (data, type, row) {
                        if (!row[9].startsWith('1970-01-01')) {
                            return '<span data-toggle="tooltip" data-placement="left" title="' + row[9] + '" data-livestamp="' + row[9] + '"></span>';
                        }
                        return row[14];
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Link
                {
                    'targets': 6,
                    'render': function (data, type, row) {
                        if (row[8]) {
                            return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    'orderable': false,
                },
            ],
        };

        // Default form inputs
        const params = new URL(window.location).searchParams;

        const $platforms = $('#platforms');
        // if (params.getAll($platforms.attr('name')).length === 0) {
        //     setUrlParam($platforms.attr('name'), getOS());
        //     $platforms.trigger("chosen:updated");
        // }

        const $appType = $('#app-type');
        // if (params.getAll($appType.attr('name')).length === 0) {
        //     setUrlParam($appType.attr('name'), ['game']);
        //     $appType.trigger("chosen:updated");
        // }

        // Init table
        const searchFields = [
            $('#search'),
            $('#tags-in'),
            $('#tags-out'),
            $('#categories'),
            $('#sale-type'),
            $appType,
            $platforms,
            $priceElement,
            $scoreElement,
            $discountElement,
            $indexElement,
        ];

        $('table.table').gdbTable({
            tableOptions: options,
            searchFields: searchFields,
        });

    })(jQuery, document);
}
