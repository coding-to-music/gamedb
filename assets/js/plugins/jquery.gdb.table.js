// the semi-colon before function invocation is a safety net against concatenated
// scripts and/or other plugins which may not be closed properly.
;(function ($, window, document, undefined) {

    "use strict";

    // Create the defaults once
    var pluginName = "gdbTable";
    var defaults = {
        fadeOnLoad: true,
        cache: true,
        tableOptions: {},
    };
    var localDefaults = {
        "search": {
            "smart": true
        },
    };
    var remoteDefaults = {
        "processing": true,
        "serverSide": true,
        "orderMulti": false,
    };

    // The actual plugin constructor
    function Plugin(element, options) {

        if (options === undefined) {
            options = {}
        }

        if (options.tableOptions === undefined) {
            options.tableOptions = {};
        }

        options.isAjax = function () {
            return this.tableOptions.columnDefs !== undefined
        }

        var tableOptions = {
            "autoWidth": false,
            "dom": '<"dt-pagination"p>t<"dt-pagination"p>r',
            "fixedHeader": true,
            "info": false,
            "language": {
                "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>',
                "paginate": {
                    "next": '<i class="fas fa-chevron-right"></i>',
                    "previous": '<i class="fas fa-chevron-left"></i>',
                },
            },
            "lengthChange": false,
            "ordering": true,
            "pageLength": 100,
            "paging": true,
            "pagingType": 'simple_numbers',
            "searching": true,
            "stateSave": false,
        }

        if (options.isAjax()) {
            tableOptions.ajax = function (data, callback, settings) {
                $.ajax({
                    url: function () {
                        const path = $(element).attr('data-path');
                        if (!path && user.log) {
                            console.log('Table data-path not set');
                        }
                        return path;
                    }(),
                    error: function (jqXHR, textStatus, errorThrown) {

                        const data = {
                            "draw": "1",
                            "recordsTotal": "0",
                            "recordsFiltered": "0",
                            "data": [],
                            "limited": false
                        };

                        callback(data, textStatus, null);
                    },
                    data: data,
                    success: callback,
                    dataType: 'json',
                    cache: true,
                });
            }
        } else {
            tableOptions.columnDefs = [
                {
                    "orderable": false,
                    "targets": $(element).find('thead tr th[data-disabled]').map(function () {
                        return $(this).index();
                    }).get(),
                }
            ]
        }

        this.settings = $.extend(true, {}, {tableOptions: tableOptions}, {tableOptions: (options.isAjax() ? remoteDefaults : localDefaults)}, options);
        this.element = element;
        this._defaults = defaults;
        this._name = pluginName;

        this.init();
    }

    $.extend(Plugin.prototype, {
        init: function () {

            this.dt = $(this.element).DataTable(this.settings.tableOptions);

            this.addDonateButton();
            this.scrollOnPaginate();
            this.hideEmptyPagination();
            this.fixImages();

            if (this.settings.isAjax()) {
                if (this.settings.fadeOnLoad) {
                    this.fadeOnLoad();
                }
            }

            // Keep track of tables
            if (window.gdbTables === undefined) {
                window.gdbTables = [];
            }
            window.gdbTables.push();

            // Fixes scrolling to pagination on every click
            $(".paginate_button > a").one("focus", function () {
                $(this).blur();
            });

            // Fixes hidden fixed header tables
            $('a[data-toggle="tab"]').one('shown.bs.tab', function (e) {
                $.each(window.gdbTables, function (index, value) {
                    value.fixedHeader.adjust();
                });
            });
        },
        fixImages: function () {
            highLightOwnedGames();
            observeLazyImages('tr img[data-lazy]');
            fixBrokenImages();
        },
        addDonateButton: function () {

            const parent = this;

            this.dt.on('xhr.dt', function (e, settings, json, xhr) {
                parent.limited = json.limited;
            });

            this.dt.on('draw.dt', function (e, settings) {
                if (parent.limited) {
                    const bold = $('li.paginate_button.page-item.next.disabled').length > 0 ? 'font-weight-bold' : '';
                    const donate = $('<li class="donate"><small><a href="/donate"><i class="fas fa-heart text-danger"></i> <span class="' + bold + '">See more!</span></a></small></li>');
                    $(parent.element).parent().find('.dt-pagination ul.pagination').append(donate);
                }
            });
        },
        fadeOnLoad: function () {
            this.dt.on('page.dt search.dt', function (e, settings) {

                $(this).fadeTo(500, 0.3);

            }).on('draw.dt', function (e, settings) {

                $(this).fadeTo(100, 1);
            });
        },
        hideEmptyPagination: function () {
            const dt = this.dt;
            dt.on('draw.dt', function (e, settings, processing) {
                if (dt.page.info().pages <= 1) {
                    $(this).parent().find('.dt-pagination').hide();
                } else {
                    $(this).parent().find('.dt-pagination').show();
                }
            });
        },
        scrollOnPaginate: function () {
            this.dt.on('page.dt', function (e, settings, processing) {

                let padding = 15;

                if ($('.fixedHeader-floating').length > 0) {
                    padding = padding + 48;
                }

                $('html, body').animate({
                    scrollTop: $(this).prev().offset().top - padding
                }, 200);
            });
        },
    });

    $.fn[pluginName] = function (options) {
        return new Plugin(this, options).dt;
    };

})(jQuery, window, document);
